package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"google.golang.org/grpc/credentials"
)

const (
	DefaultCertPath = "/etc/afp/tls/tls.crt"
	DefaultKeyPath  = "/etc/afp/tls/tls.key"
	DefaultCAPath   = "/etc/afp/tls/ca.crt"
)

// Config holds filesystem paths for AFP policy stream mTLS.
type Config struct {
	Enabled  bool
	CertPath string
	KeyPath  string
	CAPath   string
	ServerName string
}

func ConfigFromEnv() Config {
	enabled := strings.EqualFold(os.Getenv("AFP_POLICY_TLS_ENABLED"), "true")
	if raw := os.Getenv("AFP_POLICY_TLS_ENABLED"); raw == "" {
		if filesExist(DefaultCertPath, DefaultKeyPath, DefaultCAPath) {
			enabled = true
		}
	}

	serverName := os.Getenv("AFP_POLICY_TLS_SERVER_NAME")
	if serverName == "" {
		serverName = "afp-policy-controller.afp-system.svc.cluster.local"
	}

	certPath := envOr("AFP_TLS_CERT", DefaultCertPath)
	keyPath := envOr("AFP_TLS_KEY", DefaultKeyPath)
	caPath := envOr("AFP_TLS_CA", DefaultCAPath)

	return Config{
		Enabled:    enabled,
		CertPath:   certPath,
		KeyPath:    keyPath,
		CAPath:     caPath,
		ServerName: serverName,
	}
}

func ServerCredentials(cfg Config) (credentials.TransportCredentials, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	serverCert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("load server cert: %w", err)
	}

	caPEM, err := os.ReadFile(cfg.CAPath)
	if err != nil {
		return nil, fmt.Errorf("read ca cert: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("parse ca cert")
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
		MinVersion:   tls.VersionTLS12,
	}), nil
}

func ClientCredentials(cfg Config) (credentials.TransportCredentials, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	clientCert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("load client cert: %w", err)
	}

	caPEM, err := os.ReadFile(cfg.CAPath)
	if err != nil {
		return nil, fmt.Errorf("read ca cert: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("parse ca cert")
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      pool,
		ServerName:   cfg.ServerName,
		MinVersion:   tls.VersionTLS12,
	}), nil
}

func filesExist(paths ...string) bool {
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			return false
		}
	}
	return true
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
