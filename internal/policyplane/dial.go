package policyplane

import (
	"github.com/FilthyMudblood/aegis-fabric/internal/policyplane/auth"
	"github.com/FilthyMudblood/aegis-fabric/internal/policyplane/tlsconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCDialOptions assembles transport + optional SA token credentials for policy stream clients.
func GRPCDialOptions(tokenPath string) ([]grpc.DialOption, error) {
	tlsCfg := tlsconfig.ConfigFromEnv()
	var opts []grpc.DialOption

	if tlsCfg.Enabled {
		creds, err := tlsconfig.ClientCredentials(tlsCfg)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if auth.ClientAuthEnabled() {
		if tokenPath == "" {
			tokenPath = auth.DefaultSATokenPath
		}
		opts = append(opts, grpc.WithPerRPCCredentials(auth.SATokenCredentials{TokenPath: tokenPath}))
	}
	return opts, nil
}

// PolicyTLSConfig exposes the active TLS config for logging and tests.
func PolicyTLSConfig() tlsconfig.Config {
	return tlsconfig.ConfigFromEnv()
}
