package main

import (
	"encoding/binary"
	"log/slog"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/aegis-fabric/afp-sidecar/internal/dataplane"
)

func envOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func intEnvOrDefault(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func durationEnvOrDefault(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func sendOne(egressAddr, traceID string, currentDepth uint32, targetDID, payload string) error {
	conn, err := net.Dial("tcp", egressAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	depthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(depthBuf, currentDepth)
	if err := dataplane.WriteFrame(conn, []byte(traceID)); err != nil {
		return err
	}
	if err := dataplane.WriteFrame(conn, depthBuf); err != nil {
		return err
	}
	if err := dataplane.WriteFrame(conn, []byte(targetDID)); err != nil {
		return err
	}
	if err := dataplane.WriteFrame(conn, []byte(payload)); err != nil {
		return err
	}
	return nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	egressAddr := envOrDefault("AFP_EGRESS_ADDR", "127.0.0.1:8081")
	targetDID := envOrDefault("AFP_TARGET_DID", "did:afp:node:b")
	payload := envOrDefault("AFP_PAYLOAD", "burst-from-node-a")
	requests := intEnvOrDefault("AFP_REQUESTS", 500)
	interval := durationEnvOrDefault("AFP_INTERVAL", 20*time.Millisecond)

	slog.Info("starting AFP load generator",
		"egress_addr", egressAddr,
		"target_did", targetDID,
		"requests", requests,
		"interval", interval.String(),
	)

	success := 0
	fail := 0
	for i := 0; i < requests; i++ {
		traceID := "loadgen-trace-" + strconv.Itoa(i)
		if err := sendOne(egressAddr, traceID, 0, targetDID, payload); err != nil {
			fail++
			slog.Warn("loadgen request failed", "index", i, "error", err)
		} else {
			success++
		}
		time.Sleep(interval)
	}

	slog.Info("load generator finished", "success", success, "failed", fail)
}
