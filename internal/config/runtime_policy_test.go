package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRuntimePolicyHotReloadFromMountedConfigMap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AFP_ENTROPY_LIMIT")
	if err := os.WriteFile(path, []byte("0.82"), 0o644); err != nil {
		t.Fatalf("write policy file: %v", err)
	}

	rp := NewRuntimePolicy(CoreConfig{EntropyLimit: 0.95, MaxRecursionDepth: 10})
	if err := rp.WatchPolicyDir(dir); err != nil {
		t.Fatalf("watch policy dir: %v", err)
	}
	if got := rp.EntropyLimit(); got != 0.82 {
		t.Fatalf("expected hot-reloaded entropy 0.82, got %v", got)
	}
}
