package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/FilthyMudblood/aegis-fabric/internal/policyplane"
	afppolicystream "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/policystream"
	"google.golang.org/grpc"
)

func main() {
	controllerAddr := flag.String("controller", envOr("AFP_POLICY_CONTROLLER_ADDR", "127.0.0.1:8090"), "policy controller gRPC address")
	killSwitch := flag.Bool("kill-switch", false, "activate emergency kill switch")
	clearOverrides := flag.Bool("clear", false, "clear stream overrides and fall back to ConfigMap law")
	entropy := flag.Float64("entropy-limit", 0, "optional emergency entropy limit (0 = unchanged)")
	maxDepth := flag.Uint("max-recursion-depth", 0, "optional emergency recursion depth (0 = unchanged)")
	flag.Parse()

	opts, err := policyplane.GRPCDialOptions(os.Getenv("AFP_SA_TOKEN_PATH"))
	if err != nil {
		log.Fatalf("policy stream dial options: %v", err)
	}
	conn, err := grpc.NewClient(*controllerAddr, opts...)
	if err != nil {
		log.Fatalf("dial policy controller: %v", err)
	}
	defer conn.Close()

	update := &afppolicystream.PolicyUpdate{
		UpdateId:         fmt.Sprintf("policyctl-%d", time.Now().UnixNano()),
		IssuedAtUnix:     time.Now().Unix(),
		Source:           afppolicystream.PolicySource_OPERATOR_CLI,
		KillSwitchActive: *killSwitch,
		ClearOverrides:   *clearOverrides,
	}
	if *entropy > 0 {
		update.EntropyLimit = *entropy
		update.EntropyLimitSet = true
	}
	if *maxDepth > 0 {
		update.MaxRecursionDepth = uint32(*maxDepth)
		update.MaxRecursionDepthSet = true
	}

	client := afppolicystream.NewAFPPolicyStreamClient(conn)
	ack, err := client.PublishPolicyUpdate(context.Background(), update)
	if err != nil {
		log.Fatalf("publish policy update: %v", err)
	}
	fmt.Printf("accepted=%v revision=%d message=%s\n", ack.GetAccepted(), ack.GetRevision(), ack.GetMessage())
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
