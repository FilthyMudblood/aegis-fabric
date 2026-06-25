from __future__ import annotations

import functools
import logging
from typing import Any, Callable, Literal, TypeVar

from .client import AFPSidecarClient
from .exceptions import AFPQuotaExceededError

logger = logging.getLogger(__name__)

StateT = TypeVar("StateT")
NodeFn = Callable[[StateT], Any]

DEFAULT_RECURSION_DEPTH_KEY = "recursion_depth"
DEFAULT_CONTEXT_BYTES_KEY = "context_memory_bytes"
DEFAULT_BLOCKED_KEY = "afp_blocked"
DEFAULT_BLOCK_REASON_KEY = "afp_block_reason"


def _read_state_value(state: Any, key: str, default: Any = 0) -> Any:
    if isinstance(state, dict):
        return state.get(key, default)
    return getattr(state, key, default)


def _estimate_context_bytes(state: Any, key: str) -> int:
    explicit = _read_state_value(state, key, None)
    if explicit is not None:
        return int(explicit)

    messages = _read_state_value(state, "messages", None)
    if messages is None:
        return 0

    total = 0
    for message in messages:
        if isinstance(message, dict):
            total += len(str(message.get("content", "")))
        else:
            content = getattr(message, "content", message)
            total += len(str(content))
    return total


def afp_governed_node(
    fn: NodeFn | None = None,
    *,
    client: AFPSidecarClient | None = None,
    estimated_tasks: int = 1,
    target_did: str = "",
    trace_id_key: str = "trace_id",
    recursion_depth_key: str = DEFAULT_RECURSION_DEPTH_KEY,
    context_bytes_key: str = DEFAULT_CONTEXT_BYTES_KEY,
    on_quota_exceeded: Literal["raise", "annotate"] = "raise",
    blocked_key: str = DEFAULT_BLOCKED_KEY,
    block_reason_key: str = DEFAULT_BLOCK_REASON_KEY,
) -> NodeFn | Callable[[NodeFn], NodeFn]:
    """Wrap a LangGraph node with AFP pre-flight governance."""

    def decorator(node_fn: NodeFn) -> NodeFn:
        @functools.wraps(node_fn)
        def wrapper(state: StateT) -> Any:
            afp = client or AFPSidecarClient()
            recursion_depth = int(_read_state_value(state, recursion_depth_key, 0))
            context_bytes = _estimate_context_bytes(state, context_bytes_key)
            trace_id = _read_state_value(state, trace_id_key, None)

            afp.report_state(
                current_recursion_depth=recursion_depth,
                context_memory_bytes=context_bytes,
            )

            try:
                afp.pre_flight_check(
                    estimated_tasks=estimated_tasks,
                    trace_id=str(trace_id) if trace_id else None,
                    target_did=target_did,
                )
            except AFPQuotaExceededError as exc:
                logger.error("[AFP SDK] LangGraph node blocked: %s", exc)
                if on_quota_exceeded == "annotate":
                    if isinstance(state, dict):
                        return {
                            **state,
                            blocked_key: True,
                            block_reason_key: str(exc),
                        }
                    raise TypeError(
                        "annotate mode requires dict-like LangGraph state"
                    ) from exc
                raise

            return node_fn(state)

        return wrapper

    if fn is not None:
        return decorator(fn)
    return decorator


def govern_state_graph_nodes(
    graph: Any,
    node_names: list[str] | None = None,
    **kwargs: Any,
) -> Any:
    """Attach AFP governance wrappers to existing LangGraph node callables."""
    nodes = getattr(graph, "nodes", None)
    if nodes is None:
        raise TypeError("graph must expose a .nodes mapping")

    selected = node_names or list(nodes.keys())
    for name in selected:
        if name not in nodes:
            raise KeyError(f"node not found on graph: {name}")
        nodes[name] = afp_governed_node(nodes[name], **kwargs)
    return graph
