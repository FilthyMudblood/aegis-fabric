package controller

import (
	"context"
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ConfigMapReconciler struct {
	client kubernetes.Interface
}

func NewConfigMapReconciler(client kubernetes.Interface) *ConfigMapReconciler {
	return &ConfigMapReconciler{client: client}
}

func (r *ConfigMapReconciler) ReconcilePolicy(ctx context.Context, policy ClusterPolicy) ([]string, error) {
	namespaces := policy.Spec.TargetNamespaces
	if len(namespaces) == 0 {
		namespaces = []string{DefaultNamespace}
	}

	reconciled := make([]string, 0, len(namespaces))
	for _, namespace := range namespaces {
		if err := r.reconcileNamespace(ctx, namespace, policy.Spec); err != nil {
			return reconciled, fmt.Errorf("namespace %s: %w", namespace, err)
		}
		reconciled = append(reconciled, namespace)
	}
	return reconciled, nil
}

func (r *ConfigMapReconciler) reconcileNamespace(ctx context.Context, namespace string, spec ClusterPolicySpec) error {
	sidecarData := map[string]string{
		"AFP_RUN_MODE":         defaultString(spec.RunMode, "enterprise-mesh"),
		"AFP_IPC_SOCKET":       "/var/run/afp/agent.sock",
		"AFP_ENTROPY_LIMIT":    formatFloat(spec.EntropyLimit, 0.95),
		"AFP_MAX_CONTEXT_BYTES": formatUint(spec.MaxContextBytes, 536870912),
		"AFP_INGRESS_ADDR":     ":8080",
		"AFP_EGRESS_ADDR":      "127.0.0.1:8081",
		"AFP_METRICS_ADDR":     "0.0.0.0:9090",
		"AFP_MAX_RECURSION_DEPTH": formatUint32(spec.MaxRecursionDepth, 10),
	}
	agentData := map[string]string{
		"AFP_SDK_FAIL_MODE":      defaultString(spec.FailMode, "closed"),
		"AFP_IPC_SOCKET":         "/var/run/afp/agent.sock",
		"AFP_SDK_RPC_TIMEOUT_MS": "50",
	}

	if err := r.applyConfigMap(ctx, namespace, SidecarConfigMapName, sidecarData); err != nil {
		return err
	}
	return r.applyConfigMap(ctx, namespace, AgentConfigMapName, agentData)
}

func (r *ConfigMapReconciler) applyConfigMap(ctx context.Context, namespace, name string, data map[string]string) error {
	existing, err := r.client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = r.client.CoreV1().ConfigMaps(namespace).Create(ctx, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels: map[string]string{
					"app.kubernetes.io/managed-by": "afp-operator",
					"app.kubernetes.io/name":       "aegis-fabric",
				},
			},
			Data: data,
		}, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}

	if existing.Data == nil {
		existing.Data = map[string]string{}
	}
	for key, value := range data {
		existing.Data[key] = value
	}
	existing.Labels = mergeLabels(existing.Labels, map[string]string{
		"app.kubernetes.io/managed-by": "afp-operator",
		"app.kubernetes.io/name":       "aegis-fabric",
	})
	_, err = r.client.CoreV1().ConfigMaps(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	return err
}

func mergeLabels(existing, desired map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range existing {
		out[k] = v
	}
	for k, v := range desired {
		out[k] = v
	}
	return out
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func formatFloat(value, fallback float64) string {
	if value <= 0 {
		value = fallback
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func formatUint(value, fallback uint64) string {
	if value == 0 {
		value = fallback
	}
	return strconv.FormatUint(value, 10)
}

func formatUint32(value, fallback uint32) string {
	if value == 0 {
		value = fallback
	}
	return strconv.FormatUint(uint64(value), 10)
}
