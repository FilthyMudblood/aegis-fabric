package controller

import (
	"fmt"
	"os"

	"github.com/FilthyMudblood/aegis-fabric/internal/policyplane"
	afppolicystream "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/policystream"
	"google.golang.org/grpc"
)

// NewPolicyPublisherFromEnv dials the policy controller when AFP_POLICY_CONTROLLER_ADDR is set.
func NewPolicyPublisherFromEnv() (PolicyPublisher, func(), error) {
	addr := os.Getenv("AFP_POLICY_CONTROLLER_ADDR")
	if addr == "" {
		return NoopPublisher, func() {}, nil
	}

	tokenPath := os.Getenv("AFP_SA_TOKEN_PATH")
	opts, err := policyplane.GRPCDialOptions(tokenPath)
	if err != nil {
		return nil, nil, fmt.Errorf("policy stream dial options: %w", err)
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("dial policy controller: %w", err)
	}
	client := afppolicystream.NewAFPPolicyStreamClient(conn)
	return NewGRPCPolicyPublisher(client), func() { _ = conn.Close() }, nil
}
