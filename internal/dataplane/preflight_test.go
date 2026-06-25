package dataplane

import (
	"testing"

	"github.com/FilthyMudblood/aegis-fabric/internal/config"
	"github.com/FilthyMudblood/aegis-fabric/internal/core"
	afpsdk "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/sdk"
)

func newTestSEAForPreFlight(simulatedMem float64) *SingleExecutionAuthority {
	cfg := &config.SidecarConfig{
		Mode: config.ModeEnterpriseMesh,
		Core: config.CoreConfig{
			MaxToolConcurrency: 50,
			OOMPanicRatio:      0.90,
			MaxContextBytes:    1024,
		},
	}
	entropy := core.NewEntropyMonitor(&cfg.Core, &core.MockMetricsProvider{SimulatedUsage: simulatedMem})
	return NewSingleExecutionAuthority(cfg, entropy, func() uint64 { return 1 }, nil)
}

func TestEvaluatePreFlight_IsolatesOnRecursionDepth(t *testing.T) {
	sea := newTestSEAForPreFlight(0.1)
	sea.ReportInternalState(11, 0)

	resp := sea.EvaluatePreFlight("trace-loop", "", 1)
	if resp.GetAction() != afpsdk.PreFlightResponse_ISOLATED {
		t.Fatalf("expected ISOLATED, got %v", resp.GetAction())
	}
}

func TestEvaluatePreFlight_IsolatesOnIntentBurst(t *testing.T) {
	sea := newTestSEAForPreFlight(0.1)

	resp := sea.EvaluatePreFlight("trace-burst", "", 10_000)
	if resp.GetAction() != afpsdk.PreFlightResponse_ISOLATED {
		t.Fatalf("expected ISOLATED for burst, got %v (%s)", resp.GetAction(), resp.GetBlockReason())
	}
}

func TestEvaluatePreFlight_ThrottlesHighContext(t *testing.T) {
	sea := newTestSEAForPreFlight(0.1)
	sea.ReportInternalState(2, 900) // 900/1024 context pressure

	resp := sea.EvaluatePreFlight("trace-context", "", 1)
	if resp.GetAction() != afpsdk.PreFlightResponse_THROTTLED {
		t.Fatalf("expected THROTTLED, got %v (%s)", resp.GetAction(), resp.GetBlockReason())
	}
	if resp.GetDelayMs() == 0 {
		t.Fatalf("expected non-zero delay_ms")
	}
}

func TestEvaluatePreFlight_PermissiveUnderSafeLoad(t *testing.T) {
	sea := newTestSEAForPreFlight(0.1)
	sea.ReportInternalState(1, 128)

	resp := sea.EvaluatePreFlight("trace-safe", "", 1)
	if resp.GetAction() != afpsdk.PreFlightResponse_PERMISSIVE {
		t.Fatalf("expected PERMISSIVE, got %v (%s)", resp.GetAction(), resp.GetBlockReason())
	}
}
