"""Infrastructure-level exceptions raised by the AFP Python SDK."""


class AFPError(Exception):
    """Base class for AFP SDK errors."""


class AFPQuotaExceededError(AFPError):
    """Raised when the sidecar preemptively isolates intent generation."""


class AFPInfrastructureError(AFPError):
    """Raised when IPC is unreachable and fail-mode is closed."""
