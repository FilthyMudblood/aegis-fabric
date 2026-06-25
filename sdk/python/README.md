# AFP Python SDK

Lightweight IPC client for the AFP sidecar pre-flight layer.

## Install

```bash
cd sdk/python
pip install -e ".[dev]"
./scripts/gen_proto.sh
```

## Usage

```python
from afp_sdk import AFPSidecarClient, AFPQuotaExceededError

afp = AFPSidecarClient()
afp.report_state(current_recursion_depth=2, context_memory_bytes=1024 * 1024)
afp.pre_flight_check(estimated_tasks=50)
```

### LangGraph adapter

```python
from afp_sdk import afp_governed_node, AFPQuotaExceededError
from langgraph.graph import StateGraph

@afp_governed_node(estimated_tasks=10)
def planner_node(state):
    return {"recursion_depth": state["recursion_depth"] + 1}

graph = StateGraph(dict)
graph.add_node("planner", planner_node)
```

Use `on_quota_exceeded="annotate"` to write `afp_blocked` / `afp_block_reason` into state instead of raising.

### Tool decorator

```python
from afp_sdk import afp_governed

@afp_governed(estimated_tasks=1, target_did="did:afp:agent:search")
def web_search(query: str) -> str:
    ...
```

## Fail mode

- `AFP_SDK_FAIL_MODE=open` (default): IPC failures fall back to permissive behavior.
- `AFP_SDK_FAIL_MODE=closed`: IPC failures raise `AFPInfrastructureError`.

## Integration tests

```bash
cd sdk/python
pytest tests/test_ipc_integration.py -v
```
