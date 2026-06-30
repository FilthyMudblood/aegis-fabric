package controller

import (
	"context"
	"fmt"
	"time"

	afppolicystream "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/policystream"
)

// PolicyPublisher pushes declarative CRD changes onto the runtime policy stream.
type PolicyPublisher interface {
	PublishPolicy(ctx context.Context, policyName string, generation int64, spec ClusterPolicySpec) (uint64, error)
	PublishClear(ctx context.Context, policyName string, reason string) (uint64, error)
}

type noopPublisher struct{}

func (noopPublisher) PublishPolicy(context.Context, string, int64, ClusterPolicySpec) (uint64, error) {
	return 0, nil
}

func (noopPublisher) PublishClear(context.Context, string, string) (uint64, error) {
	return 0, nil
}

// NoopPublisher skips runtime stream publication (local dev without controller).
var NoopPublisher PolicyPublisher = noopPublisher{}

// GRPCPolicyPublisher calls AFPPolicyStream.PublishPolicyUpdate on the central controller.
type GRPCPolicyPublisher struct {
	client afppolicystream.AFPPolicyStreamClient
}

func NewGRPCPolicyPublisher(client afppolicystream.AFPPolicyStreamClient) *GRPCPolicyPublisher {
	return &GRPCPolicyPublisher{client: client}
}

func (p *GRPCPolicyPublisher) PublishPolicy(
	ctx context.Context,
	policyName string,
	generation int64,
	spec ClusterPolicySpec,
) (uint64, error) {
	if p == nil || p.client == nil {
		return 0, nil
	}

	entropy := spec.EntropyLimit
	if entropy <= 0 {
		entropy = 0.95
	}
	maxDepth := spec.MaxRecursionDepth
	if maxDepth == 0 {
		maxDepth = 10
	}

	ack, err := p.client.PublishPolicyUpdate(ctx, &afppolicystream.PolicyUpdate{
		UpdateId:             fmt.Sprintf("crd-%s-%d", policyName, generation),
		IssuedAtUnix:         time.Now().Unix(),
		Source:               afppolicystream.PolicySource_CRD_RECONCILER,
		KillSwitchActive:     false,
		ClearOverrides:       false,
		EntropyLimit:         entropy,
		EntropyLimitSet:      true,
		MaxRecursionDepth:    maxDepth,
		MaxRecursionDepthSet: true,
	})
	if err != nil {
		return 0, err
	}
	return ack.GetRevision(), nil
}

func (p *GRPCPolicyPublisher) PublishClear(ctx context.Context, policyName string, reason string) (uint64, error) {
	if p == nil || p.client == nil {
		return 0, nil
	}

	ack, err := p.client.PublishPolicyUpdate(ctx, &afppolicystream.PolicyUpdate{
		UpdateId:         fmt.Sprintf("crd-delete-%s-%d", policyName, time.Now().Unix()),
		IssuedAtUnix:     time.Now().Unix(),
		Source:           afppolicystream.PolicySource_CRD_RECONCILER,
		KillSwitchActive: false,
		ClearOverrides:   true,
	})
	if err != nil {
		return 0, err
	}
	_ = reason
	return ack.GetRevision(), nil
}
