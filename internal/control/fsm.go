package control

import (
	"sync"
	"sync/atomic"
	"time"
)

type State uint32

const (
	StatePermissive State = iota
	StateThrottled
	StateIsolated
	StateProbationary
)

type RoutingDecision int

const (
	ActionFastPath RoutingDecision = iota
	ActionSlowPathWithDelay
	ActionDropPacket
	ActionLowFrequencyProbe
	ActionIsolateAndBroadcast // 新增：首度触发隔离，必须广播
)

type NodeMetrics struct {
	CVPScore       float32
	EntropyLoad    float32
	CurrentEpoch   uint64
	HasValidSign   bool
	MaliciousSpike bool // 检测到死循环重试等发散行为
}

type NodeStateMachine struct {
	mu             sync.RWMutex
	nodeID         string
	state          State
	lastPenalty    uint64
	lastCVP        float32
	probationStart uint64 // 记录进入试探期的起点
}

func NewNodeStateMachine(nodeID string, initialState State) *NodeStateMachine {
	return &NodeStateMachine{
		nodeID: nodeID,
		state:  initialState,
	}
}

func (fsm *NodeStateMachine) GetState() State {
	return State(atomic.LoadUint32((*uint32)(&fsm.state)))
}

func (fsm *NodeStateMachine) EvaluateTransition(metrics *NodeMetrics) (RoutingDecision, time.Duration) {
	fsm.mu.Lock()
	defer fsm.mu.Unlock()

	currentState := fsm.state

	// Rule 1: Signature validation is the ultimate floor limit
	if !metrics.HasValidSign || metrics.MaliciousSpike || metrics.CVPScore < CVPCritical {
		return fsm.enforceIsolation(metrics)
	}

	switch currentState {
	case StatePermissive:
		// [PERMISSIVE -> THROTTLED]
		if metrics.EntropyLoad > EWarn {
			fsm.state = StateThrottled
			fsm.lastCVP = metrics.CVPScore
			return ActionSlowPathWithDelay, fsm.calculateLatency(metrics.EntropyLoad)
		}
		fsm.lastCVP = metrics.CVPScore
		return ActionFastPath, 0

	case StateThrottled:
		// [THROTTLED -> PERMISSIVE]
		if metrics.EntropyLoad < ESafe {
			fsm.state = StatePermissive
			fsm.lastCVP = metrics.CVPScore
			return ActionFastPath, 0
		}
		// Stay Throttled
		fsm.lastCVP = metrics.CVPScore
		return ActionSlowPathWithDelay, fsm.calculateLatency(metrics.EntropyLoad)

	case StateIsolated:
		epochDelta := metrics.CurrentEpoch - fsm.lastPenalty
		// [ISOLATED -> PROBATIONARY] - Exponential Backoff requirement
		if epochDelta >= 64 {
			fsm.state = StateProbationary
			fsm.probationStart = metrics.CurrentEpoch
			fsm.lastCVP = CVPCritical
			return ActionLowFrequencyProbe, 0
		}
		return ActionDropPacket, 0

	case StateProbationary:
		// Any spike in entropy during probation instantly triggers isolation
		if metrics.EntropyLoad > ESafe {
			return fsm.enforceIsolation(metrics)
		}

		deltaProbation := metrics.CurrentEpoch - fsm.probationStart
		// Only recover sufficiently if enough probation epochs have passed
		if deltaProbation > 128 && metrics.CVPScore >= 0.8 {
			fsm.state = StatePermissive
			return ActionFastPath, 0
		}
		return ActionLowFrequencyProbe, 0
	}

	return ActionDropPacket, 0
}

// enforceIsolation centralizes the drop-path logic
func (fsm *NodeStateMachine) enforceIsolation(metrics *NodeMetrics) (RoutingDecision, time.Duration) {
	wasIsolated := fsm.state == StateIsolated
	fsm.state = StateIsolated
	fsm.lastCVP = 0.0 // Ensure collateral value hits 0
	fsm.lastPenalty = metrics.CurrentEpoch

	if !wasIsolated {
		return ActionIsolateAndBroadcast, 0
	}
	return ActionDropPacket, 0
}

// calculateLatency injects 500ms to 2000ms based on how far past EWarn the load is
func (fsm *NodeStateMachine) calculateLatency(entropy float32) time.Duration {
	if entropy <= EWarn {
		return 500 * time.Millisecond
	}
	// Scale linearly from 500ms to 2000ms between EWarn(0.75) and 1.0
	excess := (entropy - EWarn) / (1.0 - EWarn)
	delayMs := 500 + int64(excess*1500)
	return time.Duration(delayMs) * time.Millisecond
}
