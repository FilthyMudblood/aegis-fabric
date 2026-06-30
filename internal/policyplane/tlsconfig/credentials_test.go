package tlsconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateDevMaterialAndLoadCredentials(t *testing.T) {
	dir := t.TempDir()
	if err := GenerateDevMaterial(dir); err != nil {
		t.Fatalf("generate dev material: %v", err)
	}

	serverCfg := Config{
		Enabled:  true,
		CertPath: filepath.Join(dir, "server.crt"),
		KeyPath:  filepath.Join(dir, "server.key"),
		CAPath:   filepath.Join(dir, "ca.crt"),
	}
	clientCfg := Config{
		Enabled:    true,
		CertPath:   filepath.Join(dir, "client.crt"),
		KeyPath:    filepath.Join(dir, "client.key"),
		CAPath:     filepath.Join(dir, "ca.crt"),
		ServerName: "afp-policy-controller.afp-system.svc.cluster.local",
	}

	serverCreds, err := ServerCredentials(serverCfg)
	if err != nil || serverCreds == nil {
		t.Fatalf("server credentials: %v", err)
	}
	clientCreds, err := ClientCredentials(clientCfg)
	if err != nil || clientCreds == nil {
		t.Fatalf("client credentials: %v", err)
	}

	for _, name := range []string{"ca.crt", "server.crt", "server.key", "client.crt", "client.key"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}

func TestConfigFromEnvDisabledWithoutFiles(t *testing.T) {
	t.Setenv("AFP_POLICY_TLS_ENABLED", "false")
	cfg := ConfigFromEnv()
	if cfg.Enabled {
		t.Fatal("expected tls disabled")
	}
}
