package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/FilthyMudblood/aegis-fabric/internal/policyplane"
	afppolicystream "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/policystream"
	"google.golang.org/grpc"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	addr := os.Getenv("AFP_POLICY_CONTROLLER_ADDR")
	if addr == "" {
		addr = ":8090"
	}

	hub := policyplane.NewHub()
	server := policyplane.NewServer(hub)
	grpcServer := grpc.NewServer()
	afppolicystream.RegisterAFPPolicyStreamServer(grpcServer, server)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("failed to bind policy controller", "error", err, "address", addr)
		os.Exit(1)
	}

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	slog.Info("AFP policy controller started", "address", addr)
	if err := grpcServer.Serve(listener); err != nil {
		slog.Error("policy controller stopped", "error", err)
		os.Exit(1)
	}
}
