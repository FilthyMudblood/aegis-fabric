package ipc

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"path/filepath"

	"github.com/FilthyMudblood/aegis-fabric/internal/dataplane"
	afpsdk "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/sdk"
	"google.golang.org/grpc"
)

// Server exposes AFPSidecarIPC over a Unix Domain Socket.
type Server struct {
	socketPath string
	sea        *dataplane.SingleExecutionAuthority
	grpcServer *grpc.Server
	listener   net.Listener
}

func NewServer(sea *dataplane.SingleExecutionAuthority, socketPath string) *Server {
	return &Server{
		socketPath: socketPath,
		sea:        sea,
	}
}

// Start binds the UDS and serves gRPC in the background.
func (s *Server) Start() error {
	if err := os.MkdirAll(filepath.Dir(s.socketPath), 0o755); err != nil {
		return err
	}
	_ = os.Remove(s.socketPath)

	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return err
	}
	s.listener = listener

	if err := os.Chmod(s.socketPath, 0o666); err != nil {
		_ = listener.Close()
		return err
	}

	s.grpcServer = grpc.NewServer()
	afpsdk.RegisterAFPSidecarIPCServer(s.grpcServer, NewPreflightService(s.sea))

	go func() {
		if err := s.grpcServer.Serve(listener); err != nil && !errors.Is(err, net.ErrClosed) {
			slog.Error("sdk ipc grpc serve error", "error", err, "socket", s.socketPath)
		}
	}()

	slog.Info("AFP SDK IPC server started", "socket", s.socketPath)
	return nil
}

func (s *Server) Serve(ctx context.Context) error {
	if s.listener == nil {
		if err := s.Start(); err != nil {
			return err
		}
	}

	<-ctx.Done()
	s.Stop()
	return nil
}

func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
	if s.listener != nil {
		_ = s.listener.Close()
	}
	if s.socketPath != "" {
		_ = os.Remove(s.socketPath)
	}
}
