package ipc_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/FilthyMudblood/aegis-fabric/internal/config"
	"github.com/FilthyMudblood/aegis-fabric/internal/core"
	"github.com/FilthyMudblood/aegis-fabric/internal/dataplane"
	"github.com/FilthyMudblood/aegis-fabric/internal/ipc"
	afpsdk "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/sdk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestIPCPreFlightOverUnixSocket(t *testing.T) {
	t.Parallel()

	socketPath := fmt.Sprintf("/tmp/afp-ipc-%d.sock", os.Getpid())
	t.Cleanup(func() { _ = os.Remove(socketPath) })
	cfg := &config.SidecarConfig{
		Mode: config.ModeEnterpriseMesh,
		Core: config.CoreConfig{
			MaxToolConcurrency: 50,
			OOMPanicRatio:      0.90,
			MaxContextBytes:    512,
		},
	}
	entropy := core.NewEntropyMonitor(&cfg.Core, &core.MockMetricsProvider{SimulatedUsage: 0.1})
	sea := dataplane.NewSingleExecutionAuthority(cfg, entropy, func() uint64 { return 1 }, nil)

	server := ipc.NewServer(sea, socketPath)
	if err := server.Start(); err != nil {
		t.Fatalf("start ipc server: %v", err)
	}
	defer server.Stop()

	conn, err := grpc.NewClient("unix://"+socketPath, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	client := afpsdk.NewAFPSidecarIPCClient(conn)
	rpcCtx, rpcCancel := context.WithTimeout(context.Background(), time.Second)
	defer rpcCancel()

	if _, err := client.ReportInternalState(rpcCtx, &afpsdk.InternalStateReport{
		CurrentRecursionDepth: 12,
		ContextMemoryBytes:    100,
	}); err != nil {
		t.Fatalf("ReportInternalState: %v", err)
	}

	resp, err := client.PreFlightCheck(rpcCtx, &afpsdk.PreFlightRequest{
		TraceId:        "ipc-test",
		EstimatedTasks: 1,
	})
	if err != nil {
		t.Fatalf("PreFlightCheck: %v", err)
	}
	if resp.GetAction() != afpsdk.PreFlightResponse_ISOLATED {
		t.Fatalf("expected ISOLATED, got %v", resp.GetAction())
	}
}
