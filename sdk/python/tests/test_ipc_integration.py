from __future__ import annotations

import os

import pytest

from afp_sdk import AFPSidecarClient
from afp_sdk.exceptions import AFPInfrastructureError, AFPQuotaExceededError

pytestmark = pytest.mark.keep_afp_singleton


def test_ipc_preflight_isolated_on_recursion_depth(sidecar_socket: str) -> None:
    AFPSidecarClient.reset_singleton()
    client = AFPSidecarClient(socket_path=sidecar_socket, fail_mode="closed")
    client.report_state(current_recursion_depth=12, context_memory_bytes=128)

    with pytest.raises(AFPQuotaExceededError):
        client.pre_flight_check(estimated_tasks=1)


def test_ipc_preflight_isolated_on_intent_burst(sidecar_socket: str) -> None:
    AFPSidecarClient.reset_singleton()
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
