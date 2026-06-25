from __future__ import annotations

from unittest.mock import Mock

import pytest

from afp_sdk.client import AFPSidecarClient
from afp_sdk.exceptions import AFPQuotaExceededError
from afp_sdk.langgraph import afp_governed_node, govern_state_graph_nodes
from afp_sdk.tools import afp_governed


@pytest.fixture(autouse=True)
def reset_client_singleton(request):
    # Keep the module-scoped sidecar client alive across IPC integration tests.
    if request.node.get_closest_marker("keep_afp_singleton"):
        yield
        return
    AFPSidecarClient.reset_singleton()
    yield
    AFPSidecarClient.reset_singleton()


def test_afp_governed_node_reports_state_and_runs():
    client = Mock(spec=AFPSidecarClient)
    client.pre_flight_check.return_value = True
    client.report_state.return_value = True

    @afp_governed_node(client=client, estimated_tasks=3)
    def planner(state):
        return {"messages": state["messages"] + ["next"]}

    result = planner(
        {
            "recursion_depth": 2,
            "context_memory_bytes": 4096,
            "messages": ["seed"],
            "trace_id": "trace-123",
        }
    )

    client.report_state.assert_called_once_with(
        current_recursion_depth=2,
        context_memory_bytes=4096,
    )
    client.pre_flight_check.assert_called_once_with(
        estimated_tasks=3,
        trace_id="trace-123",
        target_did="",
    )
    assert result["messages"] == ["seed", "next"]


def test_afp_governed_node_raises_on_quota_exceeded():
    client = Mock(spec=AFPSidecarClient)
    client.report_state.return_value = True
    client.pre_flight_check.side_effect = AFPQuotaExceededError("blocked")

    @afp_governed_node(client=client)
    def planner(state):
        return state

    with pytest.raises(AFPQuotaExceededError):
        planner({"recursion_depth": 11})


def test_afp_governed_node_annotate_mode():
    client = Mock(spec=AFPSidecarClient)
    client.report_state.return_value = True
    client.pre_flight_check.side_effect = AFPQuotaExceededError("entropy high")

    @afp_governed_node(client=client, on_quota_exceeded="annotate")
    def planner(state):
        return {"done": True}

    result = planner({"recursion_depth": 11})
    assert result["afp_blocked"] is True
    assert "entropy high" in result["afp_block_reason"]


def test_govern_state_graph_nodes_wraps_selected_nodes():
    client = Mock(spec=AFPSidecarClient)
    client.report_state.return_value = True
    client.pre_flight_check.return_value = True

    def planner(state):
        return state

    def other(state):
        return state

    graph = type("Graph", (), {"nodes": {"planner": planner, "other": other}})()
    govern_state_graph_nodes(graph, node_names=["planner"], client=client)

    graph.nodes["planner"]({"recursion_depth": 1})
    client.pre_flight_check.assert_called_once()
    client.reset_mock()
    graph.nodes["other"]({"recursion_depth": 1})
    client.pre_flight_check.assert_not_called()


def test_afp_governed_tool_decorator():
    client = Mock(spec=AFPSidecarClient)
    client.pre_flight_check.return_value = True

    @afp_governed(client=client, estimated_tasks=2, target_did="did:afp:agent:beta")
    def search_tool(query: str) -> str:
        return f"result:{query}"

    assert search_tool("weather") == "result:weather"
    client.pre_flight_check.assert_called_once_with(
        estimated_tasks=2,
        trace_id=None,
        target_did="did:afp:agent:beta",
    )
