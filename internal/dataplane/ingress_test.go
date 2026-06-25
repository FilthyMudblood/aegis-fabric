package dataplane

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/FilthyMudblood/aegis-fabric/internal/config"
	"github.com/FilthyMudblood/aegis-fabric/internal/core"
	afpv1 "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1"
)

func newTestSEA(mode config.RunMode, simulatedUsage float64) *SingleExecutionAuthority {
	cfg := &config.SidecarConfig{
		Mode: mode,
		Core: config.CoreConfig{
			MaxToolConcurrency: 50,
			MemoryWarnRatio:    0.75,
			OOMPanicRatio:      0.90,
		},
	}
	entropy := core.NewEntropyMonitor(&cfg.Core, &core.MockMetricsProvider{SimulatedUsage: simulatedUsage})
	return NewSingleExecutionAuthority(cfg, entropy, func() uint64 { return uint64(time.Now().Unix()) }, nil)
}

func TestHandleIngress_RecursionBreaker(t *testing.T) {
	sea := newTestSEA(config.ModeEnterpriseMesh, 0.1)
	err := sea.HandleIngress(context.Background(), "did:test:peer", &afpv1.GovernanceHeader{
		RecursionDepth: 11,
		CvpScore:       0.9,
		EntropyLoad:    &afpv1.EntropyLoad{ResourceAsymmetryRatio: 0.1},
	})
	if err == nil {
		t.Fatal("expected recursion breaker error, got nil")
	}
	if got := err.Error(); got != "afp-core: recursion depth exceeded physical limit, network loop detected" {
		t.Fatalf("unexpected error: %v", got)
	}
}

func TestHandleIngress_OOMCircuitBreaker(t *testing.T) {
	sea := newTestSEA(config.ModeEnterpriseMesh, 0.95)
	err := sea.HandleIngress(context.Background(), "did:test:peer", &afpv1.GovernanceHeader{
		RecursionDepth: 1,
		CvpScore:       0.9,
		EntropyLoad:    &afpv1.EntropyLoad{ResourceAsymmetryRatio: 0.1},
	})
	if err == nil {
		t.Fatal("expected circuit breaker error, got nil")
	}
	if got := err.Error(); got != "afp-core: critical context explosion or tool storm detected, circuit breaker open" {
		t.Fatalf("unexpected error: %v", got)
	}
}

func TestHandleIngress_EnterpriseBypassesStrangerTax(t *testing.T) {
	sea := newTestSEA(config.ModeEnterpriseMesh, 0.1)
	err := sea.HandleIngress(context.Background(), "did:test:anonymous", &afpv1.GovernanceHeader{
		RecursionDepth: 1,
		CvpScore:       0.9,
		EntropyLoad:    &afpv1.EntropyLoad{ResourceAsymmetryRatio: 0.1},
		// no collateral, no signature hash
	})
	if err != nil {
		t.Fatalf("expected request pass in enterprise mode, got error: %v", err)
	}
}

func TestHandleIngress_OpenExchangeRejectsStrangerWithoutCollateral(t *testing.T) {
	sea := newTestSEA(config.ModeOpenExchange, 0.1)
	err := sea.HandleIngress(context.Background(), "did:test:stranger", &afpv1.GovernanceHeader{
		RecursionDepth: 1,
		CvpScore:       0.9,
		EntropyLoad:    &afpv1.EntropyLoad{ResourceAsymmetryRatio: 0.1},
	})
	if !errors.Is(err, ErrStrangerTaxFailed) {
		t.Fatalf("expected ErrStrangerTaxFailed, got %v", err)
	}
}

func TestHandleIngress_OpenExchangeRejectsInvalidSignature(t *testing.T) {
	sea := newTestSEA(config.ModeOpenExchange, 0.1)
	err := sea.HandleIngress(context.Background(), "did:test:stranger", &afpv1.GovernanceHeader{
		RecursionDepth: 1,
		CvpScore:       0.95,
		EntropyLoad:    &afpv1.EntropyLoad{ResourceAsymmetryRatio: 0.1},
		DependencyCollateral: &afpv1.DependencyCollateral{
			CollateralType: "SYS_VIRTUAL_STAKE",
			SlashThreshold: 0.9,
		},
		// TopologyConsensusHash intentionally empty => invalid sign in open mode
	})
	if !errors.Is(err, ErrTopologyIsolated) {
		t.Fatalf("expected ErrTopologyIsolated, got %v", err)
	}
}
