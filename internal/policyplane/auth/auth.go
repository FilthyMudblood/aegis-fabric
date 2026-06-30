package auth

import (
	"context"
	"fmt"
	"os"
	"strings"

	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	MetadataAuthorization = "authorization"
	DefaultSATokenPath      = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

// Config controls Kubernetes TokenReview enforcement on the policy stream plane.
type Config struct {
	Enabled           bool
	AllowedNamespace  string
	PublisherUsername string
}

// Authenticator validates projected service account tokens via TokenReview.
type Authenticator struct {
	kube kubernetes.Interface
	cfg  Config
}

func NewAuthenticator(kube kubernetes.Interface, cfg Config) *Authenticator {
	if cfg.AllowedNamespace == "" {
		cfg.AllowedNamespace = "afp-system"
	}
	if cfg.PublisherUsername == "" {
		cfg.PublisherUsername = "system:serviceaccount:afp-system:afp-policy-operator"
	}
	return &Authenticator{kube: kube, cfg: cfg}
}

func ConfigFromEnv() Config {
	enabled := strings.EqualFold(os.Getenv("AFP_POLICY_AUTH_ENABLED"), "true")
	if raw := os.Getenv("AFP_POLICY_AUTH_ENABLED"); raw == "" {
		if _, err := os.Stat(DefaultSATokenPath); err == nil {
			if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
				enabled = true
			}
		}
	}

	namespace := os.Getenv("AFP_ALLOWED_NAMESPACE")
	if namespace == "" {
		namespace = "afp-system"
	}
	publisher := os.Getenv("AFP_PUBLISHER_SERVICE_ACCOUNT")
	if publisher == "" {
		publisher = "system:serviceaccount:afp-system:afp-policy-operator"
	}
	return Config{
		Enabled:           enabled,
		AllowedNamespace:  namespace,
		PublisherUsername: publisher,
	}
}

func (a *Authenticator) ValidateToken(ctx context.Context, token string, requirePublisher bool) error {
	if !a.cfg.Enabled {
		return nil
	}
	if a.kube == nil {
		return fmt.Errorf("kubernetes client is required when policy auth is enabled")
	}
	token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
	if token == "" {
		return status.Error(codes.Unauthenticated, "missing bearer service account token")
	}

	review, err := a.kube.AuthenticationV1().TokenReviews().Create(ctx, &authenticationv1.TokenReview{
		Spec: authenticationv1.TokenReviewSpec{Token: token},
	}, metav1.CreateOptions{})
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "token review failed: %v", err)
	}
	if !review.Status.Authenticated {
		return status.Error(codes.Unauthenticated, "service account token rejected")
	}

	username := review.Status.User.Username
	if requirePublisher {
		if username != a.cfg.PublisherUsername {
			return status.Errorf(codes.PermissionDenied, "publisher identity %q is not authorized", username)
		}
		return nil
	}

	prefix := "system:serviceaccount:" + a.cfg.AllowedNamespace + ":"
	if !strings.HasPrefix(username, prefix) {
		return status.Errorf(codes.PermissionDenied, "identity %q is outside namespace %s", username, a.cfg.AllowedNamespace)
	}
	return nil
}

func bearerTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}
	values := md.Get(MetadataAuthorization)
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization metadata")
	}
	return values[0], nil
}

func (a *Authenticator) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if a.cfg.Enabled {
			token, err := bearerTokenFromContext(ctx)
			if err != nil {
				return nil, err
			}
			requirePublisher := strings.HasSuffix(info.FullMethod, "/PublishPolicyUpdate")
			if err := a.ValidateToken(ctx, token, requirePublisher); err != nil {
				return nil, err
			}
		}
		return handler(ctx, req)
	}
}

func (a *Authenticator) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if a.cfg.Enabled {
			token, err := bearerTokenFromContext(stream.Context())
			if err != nil {
				return err
			}
			if err := a.ValidateToken(stream.Context(), token, false); err != nil {
				return err
			}
		}
		return handler(srv, stream)
	}
}

// SATokenCredentials injects a projected service account token into outbound gRPC calls.
type SATokenCredentials struct {
	TokenPath string
}

func (c SATokenCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	_ = ctx
	path := c.TokenPath
	if path == "" {
		path = DefaultSATokenPath
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	token := strings.TrimSpace(string(raw))
	if token == "" {
		return nil, fmt.Errorf("empty service account token at %s", path)
	}
	return map[string]string{MetadataAuthorization: "Bearer " + token}, nil
}

func (c SATokenCredentials) RequireTransportSecurity() bool {
	return false
}

func DialOptions(addr string, tokenPath string, authEnabled bool) []grpc.DialOption {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	if authEnabled {
		if _, err := os.Stat(firstExisting(tokenPath, DefaultSATokenPath)); err == nil {
			opts = append(opts, grpc.WithPerRPCCredentials(SATokenCredentials{TokenPath: tokenPath}))
		}
	}
	return opts
}

func firstExisting(paths ...string) string {
	for _, path := range paths {
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return DefaultSATokenPath
}
