package topology

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestResolverFastPathHit(t *testing.T) {
	store := NewInMemoryNeighborStore()
	store.UpsertEndpoint("did:afp:seed:alpha", "127.0.0.1:8080")
	resolver := NewResolver(store)

	endpoint, err := resolver.Resolve(context.Background(), "did:afp:seed:alpha")
	if err != nil {
		t.Fatalf("expected no error on fast path, got %v", err)
	}
	if endpoint != "127.0.0.1:8080" {
		t.Fatalf("unexpected endpoint: got %q", endpoint)
	}
}

func TestResolverInProgressBackoff(t *testing.T) {
	store := NewInMemoryNeighborStore()
	resolver := NewResolver(store)
	resolver.resolveCache.Store("did:afp:remote:z", true)
	defer resolver.resolveCache.Delete("did:afp:remote:z")

	_, err := resolver.Resolve(context.Background(), "did:afp:remote:z")
	if err == nil {
		t.Fatal("expected in-progress backoff error, got nil")
	}
	if got := err.Error(); got != "afp: resolution already in progress, backoff applied" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolverNoCoreNeighbors(t *testing.T) {
	store := NewInMemoryNeighborStore()
	resolver := NewResolver(store)

	_, err := resolver.Resolve(context.Background(), "did:afp:remote:z")
	if err == nil {
		t.Fatal("expected no-core-neighbors error, got nil")
	}
	if got := err.Error(); got != "afp: no core neighbors available for routing referral" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolverTimeoutForUnknownTarget(t *testing.T) {
	store := NewInMemoryNeighborStore()
	store.UpsertNeighbor("did:afp:seed:alpha", 0.95)
	store.UpsertEndpoint("did:afp:seed:alpha", "127.0.0.1:8080")
	resolver := NewResolver(store)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := resolver.Resolve(ctx, "did:afp:remote:unknown")
	if !errors.Is(err, ErrResolutionTimeout) {
		t.Fatalf("expected ErrResolutionTimeout, got %v", err)
	}
}

func TestResolverReferralAndCacheFill(t *testing.T) {
	store := NewInMemoryNeighborStore()
	store.UpsertNeighbor("did:afp:seed:alpha", 0.95)
	store.UpsertNeighbor("did:afp:seed:beta", 0.88)
	store.UpsertEndpoint("did:afp:seed:alpha", "127.0.0.1:8080")
	store.UpsertEndpoint("did:afp:seed:beta", "127.0.0.1:18080")
	resolver := NewResolver(store)

	endpoint, err := resolver.Resolve(context.Background(), "did:afp:remote:z")
	if err != nil {
		t.Fatalf("expected referral resolution success, got %v", err)
	}
	if endpoint != "127.0.0.1:8082" {
		t.Fatalf("unexpected referred endpoint: got %q", endpoint)
	}

	// second call should hit local fast-path cache populated by first referral
	fastEndpoint, fastErr := resolver.Resolve(context.Background(), "did:afp:remote:z")
	if fastErr != nil {
		t.Fatalf("expected cache fast-path success, got %v", fastErr)
	}
	if fastEndpoint != "127.0.0.1:8082" {
		t.Fatalf("unexpected cached endpoint: got %q", fastEndpoint)
	}
}
