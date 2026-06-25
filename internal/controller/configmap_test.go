package controller

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestConfigMapReconcilerCreatesSidecarAndAgentConfigMaps(t *testing.T) {
	client := fake.NewSimpleClientset()
	reconciler := NewConfigMapReconciler(client)

	policy := ClusterPolicy{
		Spec: ClusterPolicySpec{
			TargetNamespaces:  []string{"afp-system"},
			EntropyLimit:      0.88,
			MaxRecursionDepth: 8,
			RunMode:           "enterprise-mesh",
			FailMode:          "closed",
			MaxContextBytes:   1024,
		},
	}

	namespaces, err := reconciler.ReconcilePolicy(context.Background(), policy)
	if err != nil {
		t.Fatalf("reconcile policy: %v", err)
	}
	if len(namespaces) != 1 || namespaces[0] != "afp-system" {
		t.Fatalf("unexpected namespaces: %#v", namespaces)
	}

	sidecarCM, err := client.CoreV1().ConfigMaps("afp-system").Get(context.Background(), SidecarConfigMapName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get sidecar configmap: %v", err)
	}
	if sidecarCM.Data["AFP_ENTROPY_LIMIT"] != "0.88" {
		t.Fatalf("unexpected entropy limit: %s", sidecarCM.Data["AFP_ENTROPY_LIMIT"])
	}
	if sidecarCM.Data["AFP_MAX_RECURSION_DEPTH"] != "8" {
		t.Fatalf("unexpected recursion depth: %s", sidecarCM.Data["AFP_MAX_RECURSION_DEPTH"])
	}

	agentCM, err := client.CoreV1().ConfigMaps("afp-system").Get(context.Background(), AgentConfigMapName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get agent configmap: %v", err)
	}
	if agentCM.Data["AFP_SDK_FAIL_MODE"] != "closed" {
		t.Fatalf("unexpected fail mode: %s", agentCM.Data["AFP_SDK_FAIL_MODE"])
	}
}

func TestConfigMapReconcilerUpdatesExistingConfigMap(t *testing.T) {
	existing := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SidecarConfigMapName,
			Namespace: "afp-system",
		},
		Data: map[string]string{
			"AFP_ENTROPY_LIMIT": "0.99",
		},
	}
	client := fake.NewSimpleClientset(existing)
	reconciler := NewConfigMapReconciler(client)

	_, err := reconciler.ReconcilePolicy(context.Background(), ClusterPolicy{
		Spec: ClusterPolicySpec{
			TargetNamespaces:  []string{"afp-system"},
			EntropyLimit:      0.80,
			MaxRecursionDepth: 10,
			RunMode:           "enterprise-mesh",
			FailMode:          "open",
		},
	})
	if err != nil {
		t.Fatalf("reconcile policy: %v", err)
	}

	updated, err := client.CoreV1().ConfigMaps("afp-system").Get(context.Background(), SidecarConfigMapName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get updated configmap: %v", err)
	}
	if updated.Data["AFP_ENTROPY_LIMIT"] != "0.8" {
		t.Fatalf("expected updated entropy limit, got %s", updated.Data["AFP_ENTROPY_LIMIT"])
	}
}
