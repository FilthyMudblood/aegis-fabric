package main

import (
	"encoding/binary"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/FilthyMudblood/aegis-fabric/internal/dataplane"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	conn, err := net.Dial("tcp", "127.0.0.1:8081")
	if err != nil {
		slog.Error("failed to connect to local Sidecar egress", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	traceID := "trace-loop-detector-001"
	currentDepth := uint32(10) // Egress +1 -> 11, ingress should trigger recursion breaker
	targetDID := "did:afp:seed:alpha"
	payload := "loop detection payload"

	if err := dataplane.WriteFrame(conn, []byte(traceID)); err != nil {
		slog.Error("failed to write trace id frame", "error", err)
		return
	}

	depthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(depthBuf, currentDepth)
	if err := dataplane.WriteFrame(conn, depthBuf); err != nil {
		slog.Error("failed to write recursion depth frame", "error", err)
		return
	}

	if err := dataplane.WriteFrame(conn, []byte(targetDID)); err != nil {
		slog.Error("failed to write target did frame", "error", err)
		return
	}

	if err := dataplane.WriteFrame(conn, []byte(payload)); err != nil {
		slog.Error("failed to write payload frame", "error", err)
		return
	}

	slog.Info("loop test payload dispatched", "trace_id", traceID, "current_depth", currentDepth, "target_did", targetDID)
	time.Sleep(500 * time.Millisecond)
}
