from __future__ import annotations

import logging
import os
import threading
import time
import uuid
from typing import Optional

import grpc

from .exceptions import AFPInfrastructureError, AFPQuotaExceededError
from .pb import sdk_ipc_pb2, sdk_ipc_pb2_grpc

logger = logging.getLogger(__name__)

DEFAULT_SOCKET_PATH = "/var/run/afp/agent.sock"
DEFAULT_RPC_TIMEOUT_S = 0.05


class AFPSidecarClient:
    """Thread-safe singleton gRPC client for AFP sidecar pre-flight IPC."""

    _instance: Optional["AFPSidecarClient"] = None
    _instance_lock = threading.Lock()

    def __new__(cls, *args, **kwargs):
        with cls._instance_lock:
            if cls._instance is None:
                cls._instance = super().__new__(cls)
                cls._instance._initialized = False
            return cls._instance

    def __init__(
        self,
        socket_path: str | None = None,
        fail_mode: str | None = None,
        rpc_timeout_s: float | None = None,
    ) -> None:
        if getattr(self, "_initialized", False):
            return

        self.socket_path = socket_path or os.getenv("AFP_IPC_SOCKET", DEFAULT_SOCKET_PATH)
        self.fail_mode = (fail_mode or os.getenv("AFP_SDK_FAIL_MODE", "open")).strip().lower()
        if self.fail_mode not in {"open", "closed"}:
            raise ValueError("AFP_SDK_FAIL_MODE must be 'open' or 'closed'")

        timeout_env = os.getenv("AFP_SDK_RPC_TIMEOUT_MS")
        if rpc_timeout_s is not None:
            self.rpc_timeout_s = rpc_timeout_s
        elif timeout_env:
            self.rpc_timeout_s = max(float(timeout_env), 1.0) / 1000.0
        else:
            self.rpc_timeout_s = DEFAULT_RPC_TIMEOUT_S

        self._channel = grpc.insecure_channel(f"unix://{self.socket_path}")
        self._stub = sdk_ipc_pb2_grpc.AFPSidecarIPCStub(self._channel)
        self._initialized = True

    def report_state(
        self,
        current_recursion_depth: int = 0,
        context_memory_bytes: int = 0,
    ) -> bool:
        """Report application-layer state into the sidecar entropy monitor."""
        request = sdk_ipc_pb2.InternalStateReport(
            current_recursion_depth=current_recursion_depth,
            context_memory_bytes=context_memory_bytes,
        )
        try:
            response = self._stub.ReportInternalState(
                request,
                timeout=self.rpc_timeout_s,
            )
            return bool(response.received)
        except grpc.RpcError as exc:
            return self._handle_ipc_failure("ReportInternalState", exc)

    def pre_flight_check(
        self,
        estimated_tasks: int = 1,
        trace_id: str | None = None,
        target_did: str = "",
    ) -> bool:
        """Probe sidecar entropy before intent generation or outbound I/O."""
        request = sdk_ipc_pb2.PreFlightRequest(
            trace_id=trace_id or self._default_trace_id(),
            target_did=target_did,
            estimated_tasks=max(estimated_tasks, 1),
        )
        try:
            response = self._stub.PreFlightCheck(
                request,
                timeout=self.rpc_timeout_s,
            )
        except grpc.RpcError as exc:
            return self._handle_ipc_failure("PreFlightCheck", exc)

        action = response.action
        if action == sdk_ipc_pb2.PreFlightResponse.Action.PERMISSIVE:
            return True

        if action == sdk_ipc_pb2.PreFlightResponse.Action.THROTTLED:
            delay_ms = response.delay_ms or 0
            if delay_ms > 0:
                logger.warning(
                    "[AFP SDK] Throttled by infrastructure. Sleeping %sms. reason=%s",
                    delay_ms,
                    response.block_reason,
                )
                time.sleep(delay_ms / 1000.0)
            return True

        raise AFPQuotaExceededError(
            response.block_reason or "afp-core: preemptive circuit breaker open"
        )

    def close(self) -> None:
        if hasattr(self, "_channel"):
            self._channel.close()

    @classmethod
    def reset_singleton(cls) -> None:
        """Test helper to drop the cached singleton instance."""
        with cls._instance_lock:
            if cls._instance is not None:
                cls._instance.close()
            cls._instance = None

    def _default_trace_id(self) -> str:
        return f"afp-sdk-{uuid.uuid4().hex[:16]}"

    def _handle_ipc_failure(self, operation: str, exc: grpc.RpcError) -> bool:
        message = f"[AFP SDK] IPC {operation} failed: {exc.code().name} {exc.details()}"
        if self.fail_mode == "closed":
            raise AFPInfrastructureError(message) from exc

        logger.warning("%s. Falling back to network-layer enforcement (fail-open).", message)
        return True
