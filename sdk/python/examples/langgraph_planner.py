"""
Minimal LangGraph planner loop demonstrating AFP pre-flight governance.

Run with sidecar IPC available:
  AFP_IPC_SOCKET=/tmp/afp/agent.sock go run ./cmd/sidecar
  cd sdk/python && PYTHONPATH=. python examples/langgraph_planner.py

Kubernetes demo agent (looping):
  python examples/langgraph_planner.py --loop --interval 30
"""

from __future__ import annotations

import argparse
import time
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


def run_once(initial_depth: int) -> None:
    app = build_graph().compile()
    state: PlannerState = {
        "recursion_depth": initial_depth,
        "messages": ["seed"],
        "context_memory_bytes": 64,
    }

    try:
        final_state = app.invoke(state)
    except AFPQuotaExceededError as exc:
        print(f"hard-stop: {exc}", flush=True)
        return

    if final_state.get("afp_blocked"):
        print(f"annotated-stop: {final_state.get('afp_block_reason')}", flush=True)
        return

    print("completed:", final_state, flush=True)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="LangGraph planner AFP governance demo")
    parser.add_argument(
        "--loop",
        action="store_true",
        help="Re-run the planner demo on an interval (for Kubernetes log tailing)",
    )
    parser.add_argument(
        "--interval",
        type=int,
        default=30,
        help="Seconds between demo runs when --loop is set (default: 30)",
    )
    parser.add_argument(
        "--initial-depth",
        type=int,
        default=10,
        help="Starting recursion_depth; next hop should trip AFP_MAX_RECURSION_DEPTH (default: 10)",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    while True:
        print(
            f"--- langgraph planner demo (initial_depth={args.initial_depth}) ---",
            flush=True,
        )
        run_once(args.initial_depth)
        if not args.loop:
            break
        time.sleep(max(args.interval, 1))


if __name__ == "__main__":
    main()
