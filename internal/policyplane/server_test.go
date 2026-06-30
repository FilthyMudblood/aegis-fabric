package policyplane

import (
	"context"
	"net"
	"testing"
	"time"

	afppolicystream "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/policystream"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestHubPublishAndSubscribe(t *testing.T) {
	hub := NewHub()
	updates, unsubscribe := hub.Subscribe("sidecar-a")
	defer unsubscribe()

	revision := hub.Publish(&afppolicystream.PolicyUpdate{
		UpdateId:         "test-1",
		KillSwitchActive: true,
		Source:           afppolicystream.PolicySource_EMERGENCY_KILL_SWITCH,
	})

	select {
	case update := <-updates:
		if !update.GetKillSwitchActive() {
			t.Fatal("expected kill switch update")
		}
		if update.GetRevision() != revision {
			t.Fatalf("revision = %d, want %d", update.GetRevision(), revision)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for policy update")
	}
}

func TestServerStreamPolicyUpdates(t *testing.T) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)
	hub := NewHub()
	server := NewServer(hub)
	grpcServer := grpc.NewServer()
	afppolicystream.RegisterAFPPolicyStreamServer(grpcServer, server)

	go func() {
		_ = grpcServer.Serve(lis)
	}()
	t.Cleanup(grpcServer.Stop)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	client := afppolicystream.NewAFPPolicyStreamClient(conn)
	stream, err := client.StreamPolicyUpdates(ctx, &afppolicystream.PolicySubscribeRequest{
		SidecarId: "sidecar-b",
	})
	if err != nil {
		t.Fatalf("stream: %v", err)
	}

	ack, err := client.PublishPolicyUpdate(ctx, &afppolicystream.PolicyUpdate{
		KillSwitchActive: true,
		Source:           afppolicystream.PolicySource_OPERATOR_CLI,
	})
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if !ack.GetAccepted() {
		t.Fatalf("publish not accepted: %s", ack.GetMessage())
	}

	update, err := stream.Recv()
	if err != nil {
		t.Fatalf("recv: %v", err)
	}
	if !update.GetKillSwitchActive() {
		t.Fatal("expected streamed kill switch update")
	}
}
