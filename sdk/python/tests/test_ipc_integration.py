from __future__ import annotations

import os
import subprocess
import time
from pathlib import Path

import pytest

from afp_sdk import AFPSidecarClient
from afp_sdk.exceptions import AFPInfrastructureError, AFPQuotaExceededError
from afp_sdk.pb import sdk_ipc_pb2


REPO_ROOT = Path(__file__).resolve().parents[3]


@pytest.fixture(scope="module")
def sidecar_socket() -> str:
    socket_path = f"/tmp/afp-sdk-int-{os.getpid()}.sock"
    env = os.environ.copy()
    env.update(
        {
            "AFP_IPC_SOCKET": socket_path,
            "AFP_BOOTSTRAP_PATH": str(REPO_ROOT / "cmd/sidecar/bootstrap.json"),
            "AFP_INGRESS_ADDR": ":18080",
            "AFP_EGRESS_ADDR": "127.0.0.1:18081",
            "AFP_METRICS_ADDR": "127.0.0.1:19090",
        }
    )

    proc = subprocess.Popen(
        ["go", "run", "./cmd/sidecar"],
        cwd=REPO_ROOT,
        env=env,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )

    deadline = time.time() + 15
    while time.time() < deadline:
        if os.path.exists(socket_path):
            break
        if proc.poll() is not None:
            pytest.fail("sidecar process exited before IPC socket became ready")
        time.sleep(0.05)
    else:
        proc.terminate()
        pytest.fail(f"timed out waiting for IPC socket: {socket_path}")

    AFPSidecarClient.reset_singleton()
    os.environ["AFP_IPC_SOCKET"] = socket_path
    os.environ["AFP_SDK_FAIL_MODE"] = "closed"

    yield socket_path

    proc.terminate()
    proc.wait(timeout=5)
    AFPSidecarClient.reset_singleton()
    if os.path.exists(socket_path):
        os.remove(socket_path)


def test_ipc_preflight_isolated_on_recursion_depth(sidecar_socket: str) -> None:
    client = AFPSidecarClient(socket_path=sidecar_socket, fail_mode="closed")
    client.report_state(current_recursion_depth=12, context_memory_bytes=128)

    with pytest.raises(AFPQuotaExceededError):
        client.pre_flight_check(estimated_tasks=1)


def test_ipc_preflight_isolated_on_intent_burst(sidecar_socket: str) -> None:
    client = AFPSidecarClient(socket_path=sidecar_socket, fail_mode="closed")
    client.report_state(current_recursion_depth=1, context_memory_bytes=128)

    with pytest.raises(AFPQuotaExceededError):
        client.pre_flight_check(estimated_tasks=10_000)


def test_fail_open_when_sidecar_missing(monkeypatch: pytest.MonkeyPatch) -> None:
    AFPSidecarClient.reset_singleton()
    monkeypatch.setenv("AFP_SDK_FAIL_MODE", "open")
    monkeypatch.setenv("AFP_IPC_SOCKET", "/tmp/afp-sdk-missing.sock")

    client = AFPSidecarClient()
    assert client.pre_flight_check(estimated_tasks=1) is True


def test_fail_closed_when_sidecar_missing(monkeypatch: pytest.MonkeyPatch) -> None:
    AFPSidecarClient.reset_singleton()
    monkeypatch.setenv("AFP_SDK_FAIL_MODE", "closed")
    monkeypatch.setenv("AFP_IPC_SOCKET", "/tmp/afp-sdk-missing.sock")

    client = AFPSidecarClient()
    with pytest.raises(AFPInfrastructureError):
        client.pre_flight_check(estimated_tasks=1)
