package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var clusterPolicyGVR = schema.GroupVersionResource{
	Group:    Group,
	Version:  Version,
	Resource: Resource,
}

type PolicyController struct {
	dynamic    dynamic.Interface
	kube       kubernetes.Interface
	reconciler *ConfigMapReconciler
	publisher  PolicyPublisher
}

func NewPolicyController(dynamicClient dynamic.Interface, kubeClient kubernetes.Interface, publisher PolicyPublisher) *PolicyController {
	if publisher == nil {
		publisher = NoopPublisher
	}
	return &PolicyController{
		dynamic:    dynamicClient,
		kube:       kubeClient,
		reconciler: NewConfigMapReconciler(kubeClient),
		publisher:  publisher,
	}
}

func (c *PolicyController) Run(ctx context.Context) error {
	for {
		watcher, err := c.dynamic.Resource(clusterPolicyGVR).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("watch AFPClusterPolicy: %w", err)
		}
		if err := c.consumeWatch(ctx, watcher); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			slog.Warn("policy watch closed, restarting", "error", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(2 * time.Second):
		}
	}
}

func (c *PolicyController) consumeWatch(ctx context.Context, watcher watch.Interface) error {
	defer watcher.Stop()

	var wg sync.WaitGroup
	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return nil
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return fmt.Errorf("watch channel closed")
			}
			obj, ok := event.Object.(*unstructured.Unstructured)
			if !ok {
				continue
			}
			switch event.Type {
			case watch.Added, watch.Modified:
				wg.Add(1)
				go func(item *unstructured.Unstructured) {
					defer wg.Done()
					if err := c.reconcileObject(ctx, item); err != nil {
						slog.Error("policy reconcile failed", "name", item.GetName(), "error", err)
					}
				}(obj)
			case watch.Deleted:
				wg.Add(1)
				policyName := obj.GetName()
				go func(name string) {
					defer wg.Done()
					if err := c.handleDelete(ctx, name); err != nil {
						slog.Error("policy delete propagation failed", "name", name, "error", err)
					}
				}(policyName)
			}
		}
	}
}

func (c *PolicyController) reconcileObject(ctx context.Context, obj *unstructured.Unstructured) error {
	name := obj.GetName()
	generation := obj.GetGeneration()

	policy, err := decodeClusterPolicy(obj)
	if err != nil {
		_ = c.writeStatus(ctx, name, generation, StatusPhaseDegraded, err.Error(), nil, 0)
		return err
	}

	namespaces, err := c.reconciler.ReconcilePolicy(ctx, policy)
	if err != nil {
		_ = c.writeStatus(ctx, name, generation, StatusPhaseDegraded, err.Error(), nil, 0)
		return err
	}

	revision, err := c.publisher.PublishPolicy(ctx, name, generation, policy.Spec)
	if err != nil {
		msg := fmt.Sprintf("configmaps updated; stream publish failed: %v", err)
		_ = c.writeStatus(ctx, name, generation, StatusPhaseDegraded, msg, namespaces, 0)
		return fmt.Errorf("publish policy stream: %w", err)
	}

	message := fmt.Sprintf("configmaps and stream revision %d applied", revision)
	if err := c.writeStatus(ctx, name, generation, StatusPhaseApplied, message, namespaces, revision); err != nil {
		slog.Warn("status writeback failed", "name", name, "error", err)
	}

	slog.Info(
		"reconciled AFPClusterPolicy",
		"name", name,
		"generation", generation,
		"streamRevision", revision,
		"namespaces", namespaces,
		"entropyLimit", policy.Spec.EntropyLimit,
		"maxRecursionDepth", policy.Spec.MaxRecursionDepth,
	)
	return nil
}

func (c *PolicyController) handleDelete(ctx context.Context, policyName string) error {
	revision, err := c.publisher.PublishClear(ctx, policyName, "AFPClusterPolicy deleted")
	if err != nil {
		return err
	}
	slog.Info(
		"propagated AFPClusterPolicy deletion to runtime stream",
		"name", policyName,
		"streamRevision", revision,
	)
	return nil
}

func decodeClusterPolicy(obj *unstructured.Unstructured) (ClusterPolicy, error) {
	raw, err := json.Marshal(obj.Object)
	if err != nil {
		return ClusterPolicy{}, err
	}
	var policy ClusterPolicy
	if err := json.Unmarshal(raw, &policy); err != nil {
		return ClusterPolicy{}, err
	}
	return policy, nil
}
