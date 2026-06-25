"""
Minimal LangGraph planner loop demonstrating AFP pre-flight governance.

Run with sidecar IPC available:
  AFP_IPC_SOCKET=/tmp/afp/agent.sock go run ./cmd/sidecar
  cd sdk/python && PYTHONPATH=. python examples/langgraph_planner.py
"""

from __future__ import annotations

from typing import TypedDict

from langgraph.graph import END, StateGraph

from afp_sdk import AFPQuotaExceededError
from afp_sdk.langgraph import afp_governed_node


class PlannerState(TypedDict, total=False):
    recursion_depth: int
    context_memory_bytes: int
    messages: list[str]
    afp_blocked: bool
    afp_block_reason: str


@afp_governed_node(on_quota_exceeded="annotate")
def planner_node(state: PlannerState) -> PlannerState:
    depth = state.get("recursion_depth", 0) + 1
    messages = list(state.get("messages", []))
    messages.append(f"planned-step-{depth}")
    return {
        "recursion_depth": depth,
        "messages": messages,
        "context_memory_bytes": sum(len(m) for m in messages) * 1024,
    }


def build_graph() -> StateGraph:
    graph = StateGraph(PlannerState)
    graph.add_node("planner", planner_node)
    graph.set_entry_point("planner")

    def should_continue(state: PlannerState) -> str:
        if state.get("afp_blocked"):
            return "stop"
        if state.get("recursion_depth", 0) >= 12:
            return "stop"
        return "planner"

    graph.add_conditional_edges("planner", should_continue, {"planner": "planner", "stop": END})
    return graph


def main() -> None:
    app = build_graph().compile()
    state: PlannerState = {
        "recursion_depth": 10,
        "messages": ["seed"],
        "context_memory_bytes": 64,
    }

    try:
        final_state = app.invoke(state)
    except AFPQuotaExceededError as exc:
        print(f"hard-stop: {exc}")
        return

    if final_state.get("afp_blocked"):
        print(f"annotated-stop: {final_state.get('afp_block_reason')}")
        return

    print("completed:", final_state)


if __name__ == "__main__":
    main()
