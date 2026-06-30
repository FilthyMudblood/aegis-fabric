package config

import (
	afppolicystream "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/policystream"
)

// PolicyStreamOverlay is the runtime injunction layer above ConfigMap law.
type PolicyStreamOverlay struct {
	Active            bool
	KillSwitchActive  bool
	EntropyLimit      *float64
	MaxRecursionDepth *uint32
}

func overlayFromUpdate(update *afppolicystream.PolicyUpdate) *PolicyStreamOverlay {
	if update == nil {
		return nil
	}
	if update.GetClearOverrides() {
		return nil
	}

	overlay := &PolicyStreamOverlay{
		Active:           true,
		KillSwitchActive: update.GetKillSwitchActive(),
	}
	if update.GetEntropyLimitSet() {
		value := update.GetEntropyLimit()
		overlay.EntropyLimit = &value
	}
	if update.GetMaxRecursionDepthSet() {
		value := update.GetMaxRecursionDepth()
		overlay.MaxRecursionDepth = &value
	}
	return overlay
}
