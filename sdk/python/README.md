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

## Fail mode

- `AFP_SDK_FAIL_MODE=open` (default): IPC failures fall back to permissive behavior.
- `AFP_SDK_FAIL_MODE=closed`: IPC failures raise `AFPInfrastructureError`.

## Integration tests

```bash
cd sdk/python
pytest tests/test_ipc_integration.py -v
```
