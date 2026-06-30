package policyplane

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/FilthyMudblood/aegis-fabric/internal/config"
	"github.com/FilthyMudblood/aegis-fabric/internal/policyplane/auth"
	afppolicystream "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/policystream"
	"google.golang.org/grpc"
)

// StreamClient subscribes to AFPPolicyStream and applies runtime overlays.
type StreamClient struct {
	addr         string
	sidecarID    string
	namespace    string
	tokenPath    string
	authEnabled  bool
	runtime      *config.RuntimePolicy
	lastRevision atomic.Uint64
}

func NewStreamClient(addr, sidecarID, namespace string, runtime *config.RuntimePolicy) *StreamClient {
	return &StreamClient{
		addr:        addr,
		sidecarID:   sidecarID,
		namespace:   namespace,
		tokenPath:   auth.DefaultSATokenPath,
		authEnabled: auth.ClientAuthEnabled(),
		runtime:     runtime,
	}
}

func (c *StreamClient) Run(ctx context.Context) error {
	if c.addr == "" || c.runtime == nil {
		return nil
	}

	backoff := time.Second
	for {
		if err := c.consume(ctx); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slog.Warn("policy stream disconnected, retrying", "error", err, "backoff", backoff, "last_revision", c.lastRevision.Load())
		} else if ctx.Err() != nil {
			return ctx.Err()
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func (c *StreamClient) consume(ctx context.Context) error {
	opts, err := GRPCDialOptions(c.tokenPath)
	if err != nil {
		return err
	}

	conn, err := grpc.NewClient(c.addr, opts...)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := afppolicystream.NewAFPPolicyStreamClient(conn)
	stream, err := client.StreamPolicyUpdates(ctx, &afppolicystream.PolicySubscribeRequest{
		SidecarId:    c.sidecarID,
		Namespace:    c.namespace,
		LastRevision: c.lastRevision.Load(),
	})
	if err != nil {
		return err
	}

	slog.Info(
		"policy stream connected",
		"controller", c.addr,
		"sidecar_id", c.sidecarID,
		"last_revision", c.lastRevision.Load(),
		"tls_enabled", PolicyTLSConfig().Enabled,
	)
	for {
		update, err := stream.Recv()
		if err != nil {
			return err
		}
		c.lastRevision.Store(update.GetRevision())
		c.runtime.ApplyStreamUpdate(update)
		slog.Info(
			"policy stream update applied",
			"revision", update.GetRevision(),
			"kill_switch", update.GetKillSwitchActive(),
			"clear_overrides", update.GetClearOverrides(),
		)
	}
}
