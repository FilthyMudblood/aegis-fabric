package auth

import (
	"os"
	"strings"
)

func ClientAuthEnabled() bool {
	if raw := os.Getenv("AFP_POLICY_AUTH_ENABLED"); raw != "" {
		return strings.EqualFold(raw, "true")
	}
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		if _, err := os.Stat(DefaultSATokenPath); err == nil {
			return true
		}
	}
	return false
}
