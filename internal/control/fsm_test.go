package control

import "testing"

func TestFSMFastPathFromPermissive(t *testing.T) {
	fsm := NewNodeStateMachine("n1", StatePermissive)
	decision, delay := fsm.EvaluateTransition(&NodeMetrics{
		CVPScore:       0.95,
		EntropyLoad:    0.2,
		CurrentEpoch:   10,
		HasValidSign:   true,
		MaliciousSpike: false,
	})
	if decision != ActionFastPath {
		t.Fatalf("expected fast path, got %v", decision)
	}
	if delay != 0 {
		t.Fatalf("expected zero delay, got %v", delay)
	}
	if got := fsm.GetState(); got != StatePermissive {
		t.Fatalf("expected permissive state, got %v", got)
	}
}

func TestFSMThrottleAndDelay(t *testing.T) {
	fsm := NewNodeStateMachine("n1", StatePermissive)
	decision, delay := fsm.EvaluateTransition(&NodeMetrics{
		CVPScore:       0.95,
		EntropyLoad:    0.8,
		CurrentEpoch:   10,
		HasValidSign:   true,
		MaliciousSpike: false,
	})
	if decision != ActionSlowPathWithDelay {
		t.Fatalf("expected slow path, got %v", decision)
	}
	if delay <= 0 {
		t.Fatalf("expected positive delay, got %v", delay)
	}
	if got := fsm.GetState(); got != StateThrottled {
		t.Fatalf("expected throttled state, got %v", got)
	}
}

func TestFSMFirstIsolationBroadcastThenDrop(t *testing.T) {
	fsm := NewNodeStateMachine("n1", StatePermissive)
	decision1, _ := fsm.EvaluateTransition(&NodeMetrics{
		CVPScore:       0.9,
		EntropyLoad:    0.1,
		CurrentEpoch:   10,
		HasValidSign:   false,
		MaliciousSpike: false,
	})
	if decision1 != ActionIsolateAndBroadcast {
		t.Fatalf("expected isolate+broadcast on first isolation, got %v", decision1)
	}

	decision2, _ := fsm.EvaluateTransition(&NodeMetrics{
		CVPScore:       0.9,
		EntropyLoad:    0.1,
		CurrentEpoch:   11,
		HasValidSign:   false,
		MaliciousSpike: false,
	})
	if decision2 != ActionDropPacket {
		t.Fatalf("expected drop on repeated isolation, got %v", decision2)
	}
}

func TestFSMIsolatedToProbation(t *testing.T) {
	fsm := NewNodeStateMachine("n1", StateIsolated)
	fsm.lastPenalty = 1
	decision, _ := fsm.EvaluateTransition(&NodeMetrics{
		CVPScore:       0.8,
		EntropyLoad:    0.1,
		CurrentEpoch:   70,
		HasValidSign:   true,
		MaliciousSpike: false,
	})
	if decision != ActionLowFrequencyProbe {
		t.Fatalf("expected low frequency probe, got %v", decision)
	}
	if got := fsm.GetState(); got != StateProbationary {
		t.Fatalf("expected probationary state, got %v", got)
	}
}

func TestFSMProbationRecovery(t *testing.T) {
	fsm := NewNodeStateMachine("n1", StateProbationary)
	fsm.probationStart = 1
	decision, _ := fsm.EvaluateTransition(&NodeMetrics{
		CVPScore:       0.85,
		EntropyLoad:    0.2,
		CurrentEpoch:   140,
		HasValidSign:   true,
		MaliciousSpike: false,
	})
	if decision != ActionFastPath {
		t.Fatalf("expected fast path recovery, got %v", decision)
	}
	if got := fsm.GetState(); got != StatePermissive {
		t.Fatalf("expected permissive state, got %v", got)
	}
}
