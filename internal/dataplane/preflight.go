package dataplane

import (
	"fmt"

	afpcontrol "github.com/FilthyMudblood/aegis-fabric/internal/control"
	"github.com/FilthyMudblood/aegis-fabric/internal/telemetry"
	afpsdk "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/sdk"
)

const localAgentPeerID = "local-agent"

func (sea *SingleExecutionAuthority) entropyCircuitBreakerThreshold() float32 {
	limit := sea.cfg.Core.EntropyLimit
	if limit <= 0 {
		limit = 0.95
	}
	return float32(limit)
}

// ReportInternalState ingests application-layer telemetry into EntropyMonitor.
func (sea *SingleExecutionAuthority) ReportInternalState(recursionDepth uint32, contextMemoryBytes uint64) {
	sea.entropy.ReportApplicationState(recursionDepth, contextMemoryBytes)
}

// EvaluatePreFlight applies the same ACC + FSM kernel before intent leaves the agent process.
func (sea *SingleExecutionAuthority) EvaluatePreFlight(traceID, targetDID string, estimatedTasks uint32) *afpsdk.PreFlightResponse {
	_ = traceID
	_ = targetDID

	if sea.entropy.ReportedRecursionDepth() > 10 {
		telemetry.PreFlightActionTotal.WithLabelValues("isolated").Inc()
		return &afpsdk.PreFlightResponse{
			Action:      afpsdk.PreFlightResponse_ISOLATED,
			BlockReason: "afp-core: recursion depth exceeded physical limit (10), intent loop detected",
		}
	}

	realEntropy := sea.entropy.CalculatePreFlightEntropy(estimatedTasks)
	if realEntropy >= sea.entropyCircuitBreakerThreshold() {
		telemetry.PreFlightActionTotal.WithLabelValues("isolated").Inc()
		return &afpsdk.PreFlightResponse{
			Action:      afpsdk.PreFlightResponse_ISOLATED,
			BlockReason: fmt.Sprintf("afp-core: preemptive circuit breaker open (entropy=%.2f)", realEntropy),
		}
	}

	currentEpoch := sea.localEpoch()
	val, exists := sea.fsmRegistry.Load(localAgentPeerID)
	if !exists {
		newFSM := afpcontrol.NewNodeStateMachine(localAgentPeerID, afpcontrol.StatePermissive)
		actual, _ := sea.fsmRegistry.LoadOrStore(localAgentPeerID, newFSM)
		val = actual
	}

	fsm := val.(*afpcontrol.NodeStateMachine)
	metrics := &afpcontrol.NodeMetrics{
		CVPScore:       1.0,
		EntropyLoad:    realEntropy,
		CurrentEpoch:   currentEpoch,
		HasValidSign:   true,
		MaliciousSpike: estimatedTasks > uint32(sea.cfg.Core.MaxToolConcurrency),
	}

	decision, injectedDelay := fsm.EvaluateTransition(metrics)

	switch decision {
	case afpcontrol.ActionFastPath:
		telemetry.PreFlightActionTotal.WithLabelValues("permissive").Inc()
		return &afpsdk.PreFlightResponse{Action: afpsdk.PreFlightResponse_PERMISSIVE}

	case afpcontrol.ActionSlowPathWithDelay:
		telemetry.PreFlightActionTotal.WithLabelValues("throttled").Inc()
		telemetry.InjectedDelayDuration.WithLabelValues(localAgentPeerID).Observe(float64(injectedDelay.Milliseconds()))
		return &afpsdk.PreFlightResponse{
			Action:   afpsdk.PreFlightResponse_THROTTLED,
			DelayMs:  uint32(injectedDelay.Milliseconds()),
			BlockReason: fmt.Sprintf("afp-core: entropy %.2f above safe band, enforced damping", realEntropy),
		}

	case afpcontrol.ActionLowFrequencyProbe:
		telemetry.PreFlightActionTotal.WithLabelValues("throttled").Inc()
		return &afpsdk.PreFlightResponse{
			Action:   afpsdk.PreFlightResponse_THROTTLED,
			DelayMs:  1000,
			BlockReason: "afp-core: probationary probe window, intent generation damped",
		}

	default:
		telemetry.PreFlightActionTotal.WithLabelValues("isolated").Inc()
		return &afpsdk.PreFlightResponse{
			Action:      afpsdk.PreFlightResponse_ISOLATED,
			BlockReason: fmt.Sprintf("afp-core: local agent isolated (entropy=%.2f)", realEntropy),
		}
	}
}
