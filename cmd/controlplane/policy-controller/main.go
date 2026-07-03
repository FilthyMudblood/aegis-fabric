package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/FilthyMudblood/aegis-fabric/internal/policyplane"
	"github.com/FilthyMudblood/aegis-fabric/internal/policyplane/auth"
	"github.com/FilthyMudblood/aegis-fabric/internal/policyplane/tlsconfig"
	afppolicystream "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/policystream"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

	authCfg := auth.ConfigFromEnv()
	tlsCfg := tlsconfig.ConfigFromEnv()

	var serverOpts []grpc.ServerOption
	if authCfg.Enabled {
		kube, err := loadKubeClient()
		if err != nil {
			slog.Error("policy auth enabled but kubernetes client unavailable", "error", err)
			os.Exit(1)
		}
		authenticator := auth.NewAuthenticator(kube, authCfg)
		serverOpts = append(serverOpts,
			grpc.UnaryInterceptor(authenticator.UnaryServerInterceptor()),
			grpc.StreamInterceptor(authenticator.StreamServerInterceptor()),
		)
		slog.Info("policy stream authentication enabled", "namespace", authCfg.AllowedNamespace)
	}

	if tlsCfg.Enabled {
		creds, err := tlsconfig.ServerCredentials(tlsCfg)
		if err != nil {
			slog.Error("failed to load policy stream mTLS credentials", "error", err)
			os.Exit(1)
		}
		serverOpts = append(serverOpts, grpc.Creds(creds))
		slog.Info("policy stream mTLS enabled")
	}

	grpcServer := grpc.NewServer(serverOpts...)
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

	slog.Info("AFP policy controller started", "address", addr, "auth_enabled", authCfg.Enabled, "tls_enabled", tlsCfg.Enabled)
	if err := grpcServer.Serve(listener); err != nil {
		slog.Error("policy controller stopped", "error", err)
		os.Exit(1)
	}
}

func loadKubeClient() (kubernetes.Interface, error) {
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
		return kubernetes.NewForConfig(cfg)
	}
	if cfg, err := rest.InClusterConfig(); err == nil {
		return kubernetes.NewForConfig(cfg)
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}
