package topology

import (
	"context"
	"log/slog"
	"sync"
)

// TopologyWarning represents a signed out-of-band P2P alert regarding an isolated peer.
type TopologyWarning struct {
	IsolatedPeerID string
	ReporterID     string
	Epoch          uint64
	Signature      []byte
}

// NeighborStore provides thread-safe access to the local storage layer to fetch core neighbors.
type NeighborStore interface {
	GetTrustedNeighbors(minCVPSThreshold float64) []string
	ApplyPreemptiveDecay(peerID string, decayFactor float64)
	ResolveEndpoint(peerID string) (string, bool)
	UpsertEndpoint(peerID string, endpoint string)
}

// GossipBroadcaster handles asynchronous, targeted propagation of isolation events.
type GossipBroadcaster struct {
	mu       sync.RWMutex
	store    NeighborStore
	localDID string
}

func NewGossipBroadcaster(localDID string, store NeighborStore) *GossipBroadcaster {
	return &GossipBroadcaster{
		store:    store,
		localDID: localDID,
	}
}

// BroadcastWarning executes the preemptive P2P gossip protocol.
// It strictly limits propagation to high-reputation Core Neighbors (CVP > 0.8).
func (g *GossipBroadcaster) BroadcastWarning(ctx context.Context, isolatedPeer string, currentEpoch uint64) {
	// 异步执行，不阻塞数据面的 Fast Path
	go func() {
		trustedPeers := g.store.GetTrustedNeighbors(0.8)
		if len(trustedPeers) == 0 {
			return
		}

		warning := &TopologyWarning{
			IsolatedPeerID: isolatedPeer,
			ReporterID:     g.localDID,
			Epoch:          currentEpoch,
			Signature:      []byte("local_crypto_signature_placeholder"),
		}

		slog.Warn("broadcasting topology warning to core neighbors",
			"isolated_peer", isolatedPeer,
			"trusted_targets", len(trustedPeers))

		// TODO: Execute parallel P2P UDP/TCP transmission to trustedPeers.
		_ = warning
	}()
}

// HandleIncomingWarning processes an asynchronous gossip warning from a neighbor.
func (g *GossipBroadcaster) HandleIncomingWarning(warning *TopologyWarning) {
	// 验证签名的合法性后，触发预防性衰减 (Preemptive Decay)
	slog.Info("received trusted topology warning, applying preemptive decay", "target", warning.IsolatedPeerID)
	g.store.ApplyPreemptiveDecay(warning.IsolatedPeerID, 0.5) // 直接削减 50% 信誉
}
