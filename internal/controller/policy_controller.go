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
				slog.Info("cluster policy deleted", "name", obj.GetName())
			}
		}
	}
}

func (c *PolicyController) reconcileObject(ctx context.Context, obj *unstructured.Unstructured) error {
	policy, err := decodeClusterPolicy(obj)
	if err != nil {
		return err
	}
	namespaces, err := c.reconciler.ReconcilePolicy(ctx, policy)
	if err != nil {
		return err
	}
	slog.Info(
		"reconciled AFPClusterPolicy",
		"name", obj.GetName(),
		"generation", obj.GetGeneration(),
		"namespaces", namespaces,
		"entropyLimit", policy.Spec.EntropyLimit,
		"maxRecursionDepth", policy.Spec.MaxRecursionDepth,
	)
	if err := c.publisher.PublishPolicy(ctx, obj.GetName(), obj.GetGeneration(), policy.Spec); err != nil {
		return fmt.Errorf("publish policy stream: %w", err)
	}
	slog.Info("published AFPClusterPolicy to runtime stream", "name", obj.GetName(), "generation", obj.GetGeneration())
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
