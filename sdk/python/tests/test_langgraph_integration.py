from __future__ import annotations

import pytest

from afp_sdk import AFPSidecarClient
from afp_sdk.exceptions import AFPInfrastructureError, AFPQuotaExceededError
from afp_sdk.langgraph import afp_governed_node

langgraph = pytest.importorskip("langgraph")
from langgraph.graph import END, StateGraph  # noqa: E402

pytestmark = pytest.mark.keep_afp_singleton


def test_langgraph_node_isolated_by_sidecar(sidecar_socket: str) -> None:
    AFPSidecarClient.reset_singleton()

    @afp_governed_node(on_quota_exceeded="annotate")
    def planner(state: dict) -> dict:
        return {
            **state,
            "recursion_depth": state.get("recursion_depth", 0) + 1,
        }

    graph = StateGraph(dict)
    graph.add_node("planner", planner)
    graph.set_entry_point("planner")
    graph.add_edge("planner", END)
    app = graph.compile()

    result = app.invoke(
        {
            "recursion_depth": 12,
            "context_memory_bytes": 128,
        }
    )

    assert result["afp_blocked"] is True
    assert "recursion depth" in result["afp_block_reason"].lower()
