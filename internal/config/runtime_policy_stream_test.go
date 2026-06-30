package config

import (
	"testing"

	afppolicystream "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/policystream"
)

func TestRuntimePolicyStreamOverlayMerge(t *testing.T) {
	rp := NewRuntimePolicy(CoreConfig{EntropyLimit: 0.95, MaxRecursionDepth: 10})

	rp.ApplyStreamUpdate(&afppolicystream.PolicyUpdate{
		KillSwitchActive:     false,
		EntropyLimit:         0.80,
		EntropyLimitSet:      true,
		MaxRecursionDepth:    6,
		MaxRecursionDepthSet: true,
	})

	if got := rp.EntropyLimit(); got != 0.80 {
		t.Fatalf("entropy limit = %v, want 0.80", got)
	}
	if got := rp.MaxRecursionDepth(); got != 6 {
		t.Fatalf("max recursion depth = %d, want 6", got)
	}
	if rp.KillSwitchActive() {
		t.Fatal("kill switch should be inactive")
	}
}

func TestRuntimePolicyKillSwitch(t *testing.T) {
	rp := NewRuntimePolicy(CoreConfig{EntropyLimit: 0.95, MaxRecursionDepth: 10})
	rp.ApplyStreamUpdate(&afppolicystream.PolicyUpdate{KillSwitchActive: true})

	if !rp.KillSwitchActive() {
		t.Fatal("expected kill switch active")
	}
}

func TestRuntimePolicyClearOverridesFallsBackToConfigMapLaw(t *testing.T) {
	rp := NewRuntimePolicy(CoreConfig{EntropyLimit: 0.95, MaxRecursionDepth: 10})
	rp.ApplyStreamUpdate(&afppolicystream.PolicyUpdate{
		EntropyLimit:         0.50,
		EntropyLimitSet:      true,
		MaxRecursionDepth:    3,
		MaxRecursionDepthSet: true,
	})
	rp.ApplyStreamUpdate(&afppolicystream.PolicyUpdate{ClearOverrides: true})

	if got := rp.EntropyLimit(); got != 0.95 {
		t.Fatalf("entropy limit = %v, want base 0.95", got)
	}
	if got := rp.MaxRecursionDepth(); got != 10 {
		t.Fatalf("max recursion depth = %d, want base 10", got)
	}
}

func TestRuntimePolicyConfigMapStillUpdatesBaseLayer(t *testing.T) {
	rp := NewRuntimePolicy(CoreConfig{EntropyLimit: 0.95, MaxRecursionDepth: 10})
	rp.ApplyEnvMap(map[string]string{
		"AFP_ENTROPY_LIMIT":       "0.88",
		"AFP_MAX_RECURSION_DEPTH": "8",
	})

	if got := rp.EntropyLimit(); got != 0.88 {
		t.Fatalf("entropy limit = %v, want 0.88", got)
	}
	if got := rp.MaxRecursionDepth(); got != 8 {
		t.Fatalf("max recursion depth = %d, want 8", got)
	}
}
