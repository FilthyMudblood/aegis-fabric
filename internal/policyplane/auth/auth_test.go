package auth

import (
	"context"
	"testing"

	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestAuthenticatorRejectsMissingToken(t *testing.T) {
	auth := NewAuthenticator(fake.NewSimpleClientset(), Config{Enabled: true})
	err := auth.ValidateToken(context.Background(), "", false)
	if err == nil {
		t.Fatal("expected missing token error")
	}
}

func TestAuthenticatorAcceptsSidecarIdentity(t *testing.T) {
	client := fake.NewSimpleClientset()
	client.PrependReactor("create", "tokenreviews", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, &authenticationv1.TokenReview{
			Status: authenticationv1.TokenReviewStatus{
				Authenticated: true,
				User: authenticationv1.UserInfo{
					Username: "system:serviceaccount:afp-system:afp-sidecar-agent",
				},
			},
		}, nil
	})

	auth := NewAuthenticator(client, Config{
		Enabled:          true,
		AllowedNamespace: "afp-system",
	})
	if err := auth.ValidateToken(context.Background(), "token", false); err != nil {
		t.Fatalf("expected sidecar identity to pass: %v", err)
	}
}

func TestAuthenticatorRequiresPublisherForPublishPath(t *testing.T) {
	client := fake.NewSimpleClientset()
	client.PrependReactor("create", "tokenreviews", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, &authenticationv1.TokenReview{
			Status: authenticationv1.TokenReviewStatus{
				Authenticated: true,
				User: authenticationv1.UserInfo{
					Username: "system:serviceaccount:afp-system:afp-sidecar-agent",
				},
			},
		}, nil
	})

	auth := NewAuthenticator(client, Config{
		Enabled:           true,
		AllowedNamespace:  "afp-system",
		PublisherUsername: "system:serviceaccount:afp-system:afp-policy-operator",
	})
	if err := auth.ValidateToken(context.Background(), "token", true); err == nil {
		t.Fatal("expected publisher-only rejection")
	}
}

func TestAuthenticatorAcceptsPublisherIdentity(t *testing.T) {
	client := fake.NewSimpleClientset()
	client.PrependReactor("create", "tokenreviews", func(action k8stesting.Action) (bool, runtime.Object, error) {
		review := action.(k8stesting.CreateAction).GetObject().(*authenticationv1.TokenReview)
		if review.Spec.Token != "operator-token" {
			t.Fatalf("unexpected token: %s", review.Spec.Token)
		}
		return true, &authenticationv1.TokenReview{
			Status: authenticationv1.TokenReviewStatus{
				Authenticated: true,
				User: authenticationv1.UserInfo{
					Username: "system:serviceaccount:afp-system:afp-policy-operator",
				},
			},
		}, nil
	})

	auth := NewAuthenticator(client, Config{
		Enabled:           true,
		PublisherUsername: "system:serviceaccount:afp-system:afp-policy-operator",
	})
	if err := auth.ValidateToken(context.Background(), "operator-token", true); err != nil {
		t.Fatalf("expected publisher identity to pass: %v", err)
	}
}

func TestConfigFromEnvDefaults(t *testing.T) {
	t.Setenv("AFP_POLICY_AUTH_ENABLED", "true")
	cfg := ConfigFromEnv()
	if !cfg.Enabled || cfg.AllowedNamespace != "afp-system" {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}
