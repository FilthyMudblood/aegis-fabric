from __future__ import annotations

import os
import socket
import subprocess
import time
from pathlib import Path

import grpc
import pytest

from afp_sdk import AFPSidecarClient
from afp_sdk.pb import sdk_ipc_pb2, sdk_ipc_pb2_grpc

REPO_ROOT = Path(__file__).resolve().parents[3]


def _pick_loopback_port() -> int:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        sock.bind(("127.0.0.1", 0))
        return int(sock.getsockname()[1])


def _wait_for_ipc(socket_path: str, proc: subprocess.Popen[str], timeout_s: float = 20.0) -> None:
    deadline = time.time() + timeout_s
    target = f"unix://{socket_path}"

    while time.time() < deadline:
        if proc.poll() is not None:
            pytest.fail("sidecar process exited before IPC socket became ready")

        if not os.path.exists(socket_path):
            time.sleep(0.05)
            continue

        channel = grpc.insecure_channel(target)
        stub = sdk_ipc_pb2_grpc.AFPSidecarIPCStub(channel)
        try:
            stub.ReportInternalState(
                sdk_ipc_pb2.InternalStateReport(),
                timeout=0.2,
            )
            channel.close()
            return
        except grpc.RpcError:
            channel.close()
            time.sleep(0.05)

    proc.terminate()
    pytest.fail(f"timed out waiting for IPC readiness: {socket_path}")


@pytest.fixture(scope="module")
def sidecar_socket() -> str:
    socket_path = f"/tmp/afp-sdk-int-{os.getpid()}.sock"
    if os.path.exists(socket_path):
        os.remove(socket_path)

    ingress_port = _pick_loopback_port()
    egress_port = _pick_loopback_port()
    metrics_port = _pick_loopback_port()

    env = os.environ.copy()
    env.update(
        {
            "AFP_IPC_SOCKET": socket_path,
            "AFP_BOOTSTRAP_PATH": str(REPO_ROOT / "cmd/sidecar/bootstrap.json"),
            "AFP_INGRESS_ADDR": f":{ingress_port}",
            "AFP_EGRESS_ADDR": f"127.0.0.1:{egress_port}",
            "AFP_METRICS_ADDR": f"127.0.0.1:{metrics_port}",
        }
    )

    proc = subprocess.Popen(
        ["go", "run", "./cmd/sidecar"],
        cwd=REPO_ROOT,
        env=env,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )

    try:
        _wait_for_ipc(socket_path, proc)
    except Exception:
        proc.terminate()
        proc.wait(timeout=5)
        raise

    AFPSidecarClient.reset_singleton()
    os.environ["AFP_IPC_SOCKET"] = socket_path
    os.environ["AFP_SDK_FAIL_MODE"] = "closed"

    yield socket_path

    proc.terminate()
    proc.wait(timeout=5)
    AFPSidecarClient.reset_singleton()
    if os.path.exists(socket_path):
        os.remove(socket_path)
