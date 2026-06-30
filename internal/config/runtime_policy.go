package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	afppolicystream "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/policystream"
)

// RuntimePolicy holds hot-reloadable sidecar thresholds.
// ConfigMap/fsnotify is declarative law; stream overlay is runtime injunction.
type RuntimePolicy struct {
	baseEntropy      atomic.Value // float64
	baseMaxDepth     atomic.Uint32
	entropyLimit     atomic.Value // float64 effective
	maxRecursionDepth atomic.Uint32
	killSwitch       atomic.Bool

	overlayMu sync.RWMutex
	overlay   *PolicyStreamOverlay
}

func NewRuntimePolicy(base CoreConfig) *RuntimePolicy {
	rp := &RuntimePolicy{}
	if base.MaxRecursionDepth == 0 {
		base.MaxRecursionDepth = 10
	}
	if base.EntropyLimit <= 0 {
		base.EntropyLimit = 0.95
	}
	rp.baseEntropy.Store(base.EntropyLimit)
	rp.baseMaxDepth.Store(base.MaxRecursionDepth)
	rp.recomputeEffective()
	return rp
}

func (rp *RuntimePolicy) EntropyLimit() float32 {
	value, ok := rp.entropyLimit.Load().(float64)
	if !ok || value <= 0 {
		return 0.95
	}
	return float32(value)
}

func (rp *RuntimePolicy) MaxRecursionDepth() uint32 {
	value := rp.maxRecursionDepth.Load()
	if value == 0 {
		return 10
	}
	return value
}

func (rp *RuntimePolicy) KillSwitchActive() bool {
	return rp.killSwitch.Load()
}

func (rp *RuntimePolicy) ApplyStreamUpdate(update *afppolicystream.PolicyUpdate) {
	rp.overlayMu.Lock()
	defer rp.overlayMu.Unlock()
	rp.overlay = overlayFromUpdate(update)
	rp.recomputeEffective()
}

func (rp *RuntimePolicy) StreamOverlay() *PolicyStreamOverlay {
	rp.overlayMu.RLock()
	defer rp.overlayMu.RUnlock()
	if rp.overlay == nil {
		return nil
	}
	copy := *rp.overlay
	return &copy
}

func (rp *RuntimePolicy) recomputeEffective() {
	baseEntropy, _ := rp.baseEntropy.Load().(float64)
	if baseEntropy <= 0 {
		baseEntropy = 0.95
	}
	baseDepth := rp.baseMaxDepth.Load()
	if baseDepth == 0 {
		baseDepth = 10
	}

	effectiveEntropy := baseEntropy
	effectiveDepth := baseDepth
	killSwitch := false

	if rp.overlay != nil && rp.overlay.Active {
		killSwitch = rp.overlay.KillSwitchActive
		if rp.overlay.EntropyLimit != nil {
			effectiveEntropy = *rp.overlay.EntropyLimit
		}
		if rp.overlay.MaxRecursionDepth != nil {
			effectiveDepth = *rp.overlay.MaxRecursionDepth
		}
	}

	rp.entropyLimit.Store(effectiveEntropy)
	rp.maxRecursionDepth.Store(effectiveDepth)
	rp.killSwitch.Store(killSwitch)
}

func (rp *RuntimePolicy) ApplyEnvMap(data map[string]string) {
	rp.overlayMu.Lock()
	defer rp.overlayMu.Unlock()
	if raw, ok := data["AFP_ENTROPY_LIMIT"]; ok {
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(raw), 64); err == nil && parsed > 0 {
			rp.baseEntropy.Store(parsed)
		}
	}
	if raw, ok := data["AFP_MAX_RECURSION_DEPTH"]; ok {
		if parsed, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 32); err == nil && parsed > 0 {
			rp.baseMaxDepth.Store(uint32(parsed))
		}
	}
	rp.recomputeEffective()
}

func (rp *RuntimePolicy) WatchPolicyDir(dir string) error {
	if dir == "" {
		return nil
	}
	if err := rp.loadDir(dir); err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := watcher.Add(dir); err != nil {
		_ = watcher.Close()
		return err
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {
					_ = rp.loadDir(dir)
				}
			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()
	return nil
}

func (rp *RuntimePolicy) loadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	data := map[string]string{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := strings.TrimSpace(string(raw))
		if strings.Contains(content, "=") {
			fileData, err := parseEnvFile(path)
			if err != nil {
				continue
			}
			for key, value := range fileData {
				data[key] = value
			}
			continue
		}
		data[entry.Name()] = content
	}
	if len(data) > 0 {
		rp.ApplyEnvMap(data)
	}
	return nil
}

func parseEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		data[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return data, scanner.Err()
}
