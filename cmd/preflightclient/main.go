package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	afpsdk "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/sdk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	socketPath := flag.String("socket", envOrDefault("AFP_IPC_SOCKET", "/var/run/afp/agent.sock"), "Unix domain socket path for AFP SDK IPC")
	traceID := flag.String("trace-id", "preflight-smoke", "OpenTelemetry-compatible trace id")
	targetDID := flag.String("target-did", "", "Optional target agent DID")
	estimatedTasks := flag.Uint("estimated-tasks", 1, "Estimated child tasks about to be generated")
	recursionDepth := flag.Uint("recursion-depth", 0, "Optional application recursion depth to report before check")
	contextBytes := flag.Uint64("context-bytes", 0, "Optional application context size in bytes to report before check")
	timeout := flag.Duration("timeout", 50*time.Millisecond, "RPC timeout")
	flag.Parse()

	addr := "unix://" + *socketPath
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "dial failed: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	client := afpsdk.NewAFPSidecarIPCClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	if *recursionDepth > 0 || *contextBytes > 0 {
		ack, err := client.ReportInternalState(ctx, &afpsdk.InternalStateReport{
			CurrentRecursionDepth: uint32(*recursionDepth),
			ContextMemoryBytes:    *contextBytes,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "ReportInternalState failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("reported internal state: received=%v\n", ack.GetReceived())
	}

	resp, err := client.PreFlightCheck(ctx, &afpsdk.PreFlightRequest{
		TraceId:        *traceID,
		TargetDid:      *targetDID,
		EstimatedTasks: uint32(*estimatedTasks),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "PreFlightCheck failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("action=%s delay_ms=%d reason=%q\n", resp.GetAction().String(), resp.GetDelayMs(), resp.GetBlockReason())
	switch resp.GetAction() {
	case afpsdk.PreFlightResponse_ISOLATED:
		os.Exit(2)
	case afpsdk.PreFlightResponse_THROTTLED:
		os.Exit(0)
	default:
		os.Exit(0)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
