package config

import "testing"

func TestLoadEnvConfig_EntropyLimit(t *testing.T) {
	t.Setenv("AFP_ENTROPY_LIMIT", "0.88")
	cfg := LoadEnvConfig()
	if cfg.Core.EntropyLimit != 0.88 {
		t.Fatalf("expected entropy limit 0.88, got %v", cfg.Core.EntropyLimit)
	}
}
