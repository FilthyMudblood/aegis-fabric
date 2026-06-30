package policyplane

import (
	"sync"

	afppolicystream "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/policystream"
)

type subscriber struct {
	id      string
	updates chan *afppolicystream.PolicyUpdate
}

// Hub stores the latest runtime policy and fans out updates to sidecar subscribers.
type Hub struct {
	mu       sync.RWMutex
	revision uint64
	current  *afppolicystream.PolicyUpdate
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
	copy := *h.current
	return &copy, h.revision
}

func (h *Hub) Publish(update *afppolicystream.PolicyUpdate) uint64 {
	if update == nil {
		return h.revision
	}

	h.mu.Lock()
	h.revision++
	update.Revision = h.revision
	h.current = update
	subs := make([]*subscriber, 0, len(h.subs))
	for _, sub := range h.subs {
		subs = append(subs, sub)
	}
	h.mu.Unlock()

	for _, sub := range subs {
		select {
		case sub.updates <- update:
		default:
		}
	}
	return h.revision
}

func (h *Hub) Subscribe(sidecarID string) (<-chan *afppolicystream.PolicyUpdate, func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if existing, ok := h.subs[sidecarID]; ok {
		close(existing.updates)
	}

	sub := &subscriber{
		id:      sidecarID,
		updates: make(chan *afppolicystream.PolicyUpdate, 8),
	}
	h.subs[sidecarID] = sub

	if h.current != nil {
		snapshot := *h.current
		sub.updates <- &snapshot
	}

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
