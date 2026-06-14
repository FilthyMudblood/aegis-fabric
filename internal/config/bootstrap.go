package config

import (
	"encoding/json"
	"os"
)

// SeedNode defines a statically trusted peer used for network cold starts.
type SeedNode struct {
	PeerID   string  `json:"peer_id"`
	CVP      float64 `json:"cvp"`
	Endpoint string  `json:"endpoint,omitempty"`
}

// BootstrapConfig holds the Genesis topology state.
type BootstrapConfig struct {
	SeedNodes []SeedNode `json:"seed_nodes"`
}

// LoadBootstrapConfig strictly parses the genesis JSON.
// Fails fast if the file is missing or malformed to prevent isolated starts.
func LoadBootstrapConfig(path string) (*BootstrapConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg BootstrapConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
