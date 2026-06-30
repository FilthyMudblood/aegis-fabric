package controller

import (
	"context"
	"fmt"
	"time"

	afppolicystream "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/policystream"
)

// PolicyPublisher pushes declarative CRD changes onto the runtime policy stream.
type PolicyPublisher interface {
	PublishPolicy(ctx context.Context, policyName string, generation int64, spec ClusterPolicySpec) error
}

type noopPublisher struct{}

func (noopPublisher) PublishPolicy(context.Context, string, int64, ClusterPolicySpec) error {
	return nil
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
) error {
	if p == nil || p.client == nil {
		return nil
	}

	entropy := spec.EntropyLimit
	if entropy <= 0 {
		entropy = 0.95
	}
	maxDepth := spec.MaxRecursionDepth
	if maxDepth == 0 {
		maxDepth = 10
	}

	_, err := p.client.PublishPolicyUpdate(ctx, &afppolicystream.PolicyUpdate{
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
	return err
}
