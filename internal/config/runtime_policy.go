package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
)

// RuntimePolicy holds hot-reloadable sidecar thresholds.
type RuntimePolicy struct {
	entropyLimit      atomic.Value // float64
	maxRecursionDepth atomic.Uint32
}

func NewRuntimePolicy(base CoreConfig) *RuntimePolicy {
	rp := &RuntimePolicy{}
	rp.entropyLimit.Store(base.EntropyLimit)
	if base.MaxRecursionDepth == 0 {
		base.MaxRecursionDepth = 10
	}
	rp.maxRecursionDepth.Store(base.MaxRecursionDepth)
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

func (rp *RuntimePolicy) ApplyEnvMap(data map[string]string) {
	if raw, ok := data["AFP_ENTROPY_LIMIT"]; ok {
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(raw), 64); err == nil && parsed > 0 {
			rp.entropyLimit.Store(parsed)
		}
	}
	if raw, ok := data["AFP_MAX_RECURSION_DEPTH"]; ok {
		if parsed, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 32); err == nil && parsed > 0 {
			rp.maxRecursionDepth.Store(uint32(parsed))
		}
	}
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
