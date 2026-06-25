from __future__ import annotations

import functools
from typing import Any, Callable, TypeVar

from .client import AFPSidecarClient

Fn = TypeVar("Fn", bound=Callable[..., Any])


def afp_governed(
    fn: Fn | None = None,
    *,
    client: AFPSidecarClient | None = None,
    estimated_tasks: int = 1,
    target_did: str = "",
    trace_id: str | None = None,
) -> Fn | Callable[[Fn], Fn]:
    """Decorator for tool/agent callables that may trigger outbound I/O."""

    def decorator(tool_fn: Fn) -> Fn:
        @functools.wraps(tool_fn)
        def wrapper(*args: Any, **kwargs: Any) -> Any:
            afp = client or AFPSidecarClient()
            afp.pre_flight_check(
                estimated_tasks=estimated_tasks,
                trace_id=trace_id,
                target_did=target_did,
            )
            return tool_fn(*args, **kwargs)

        return wrapper  # type: ignore[return-value]

    if fn is not None:
        return decorator(fn)
    return decorator
