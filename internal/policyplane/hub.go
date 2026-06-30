package policyplane

import (
	"sync"

	afppolicystream "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/policystream"
	"google.golang.org/protobuf/proto"
)

const defaultHistoryLimit = 128

type subscriber struct {
	id      string
	updates chan *afppolicystream.PolicyUpdate
}

// Hub stores the latest runtime policy and fans out updates to sidecar subscribers.
type Hub struct {
	mu       sync.RWMutex
	revision uint64
	current  *afppolicystream.PolicyUpdate
	history  []*afppolicystream.PolicyUpdate
	subs     map[string]*subscriber
}

func NewHub() *Hub {
	return &Hub{
		subs: make(map[string]*subscriber),
	}
}

func (h *Hub) Current() (*afppolicystream.PolicyUpdate, uint64) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.current == nil {
		return nil, h.revision
	}
	copy := proto.Clone(h.current).(*afppolicystream.PolicyUpdate)
	return copy, h.revision
}

func (h *Hub) HistorySince(lastRevision uint64) []*afppolicystream.PolicyUpdate {
	h.mu.RLock()
	defer h.mu.RUnlock()

	out := make([]*afppolicystream.PolicyUpdate, 0)
	for _, update := range h.history {
		if update.GetRevision() > lastRevision {
			out = append(out, proto.Clone(update).(*afppolicystream.PolicyUpdate))
		}
	}
	return out
}

func (h *Hub) Publish(update *afppolicystream.PolicyUpdate) uint64 {
	if update == nil {
		return h.revision
	}

	h.mu.Lock()
	h.revision++
	update.Revision = h.revision
	stored := proto.Clone(update).(*afppolicystream.PolicyUpdate)
	h.current = stored
	h.history = append(h.history, stored)
	if len(h.history) > defaultHistoryLimit {
		h.history = append([]*afppolicystream.PolicyUpdate(nil), h.history[len(h.history)-defaultHistoryLimit:]...)
	}
	subs := make([]*subscriber, 0, len(h.subs))
	for _, sub := range h.subs {
		subs = append(subs, sub)
	}
	h.mu.Unlock()

	for _, sub := range subs {
		select {
		case sub.updates <- proto.Clone(stored).(*afppolicystream.PolicyUpdate):
		default:
		}
	}
	return h.revision
}

func (h *Hub) Subscribe(sidecarID string, lastRevision uint64) (<-chan *afppolicystream.PolicyUpdate, func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if existing, ok := h.subs[sidecarID]; ok {
		close(existing.updates)
	}

	sub := &subscriber{
		id:      sidecarID,
		updates: make(chan *afppolicystream.PolicyUpdate, 16),
	}
	h.subs[sidecarID] = sub
	h.replayLocked(sub, lastRevision)

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if current, ok := h.subs[sidecarID]; ok && current == sub {
			delete(h.subs, sidecarID)
			close(sub.updates)
		}
	}
	return sub.updates, unsubscribe
}

func (h *Hub) replayLocked(sub *subscriber, lastRevision uint64) {
	if lastRevision == 0 {
		if h.current != nil {
			sub.updates <- proto.Clone(h.current).(*afppolicystream.PolicyUpdate)
		}
		return
	}
	for _, update := range h.history {
		if update.GetRevision() > lastRevision {
			sub.updates <- proto.Clone(update).(*afppolicystream.PolicyUpdate)
		}
	}
}
