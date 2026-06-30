package controller

import (
	"context"
	"net"
	"testing"

	"github.com/FilthyMudblood/aegis-fabric/internal/policyplane"
	afppolicystream "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/policystream"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestGRPCPolicyPublisherPublishesCRDSpec(t *testing.T) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)
	hub := policyplane.NewHub()
	server := policyplane.NewServer(hub)
	grpcServer := grpc.NewServer()
	afppolicystream.RegisterAFPPolicyStreamServer(grpcServer, server)
	go func() { _ = grpcServer.Serve(lis) }()
	t.Cleanup(grpcServer.Stop)

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	publisher := NewGRPCPolicyPublisher(afppolicystream.NewAFPPolicyStreamClient(conn))
	err = publisher.PublishPolicy(context.Background(), "enterprise-default", 3, ClusterPolicySpec{
		EntropyLimit:      0.82,
		MaxRecursionDepth: 7,
	})
	if err != nil {
		t.Fatalf("publish policy: %v", err)
	}

	current, revision := hub.Current()
	if revision != 1 {
		t.Fatalf("revision = %d, want 1", revision)
	}
	if current.GetEntropyLimit() != 0.82 {
		t.Fatalf("entropy = %v", current.GetEntropyLimit())
	}
	if current.GetMaxRecursionDepth() != 7 {
		t.Fatalf("depth = %d", current.GetMaxRecursionDepth())
	}
	if current.GetSource() != afppolicystream.PolicySource_CRD_RECONCILER {
		t.Fatalf("source = %v", current.GetSource())
	}
}
