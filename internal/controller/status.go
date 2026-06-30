package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	StatusPhaseApplied  = "Applied"
	StatusPhaseDegraded = "Degraded"
)

type clusterPolicyStatus struct {
	ObservedGeneration   string   `json:"observedGeneration,omitempty"`
	Phase                string   `json:"phase,omitempty"`
	Message              string   `json:"message,omitempty"`
	ReconciledNamespaces []string `json:"reconciledNamespaces,omitempty"`
	StreamRevision       string   `json:"streamRevision,omitempty"`
}

func (c *PolicyController) writeStatus(
	ctx context.Context,
	name string,
	generation int64,
	phase string,
	message string,
	namespaces []string,
	streamRevision uint64,
) error {
	current, err := c.dynamic.Resource(clusterPolicyGVR).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get policy for status write: %w", err)
	}

	status := clusterPolicyStatus{
		ObservedGeneration:   strconv.FormatInt(generation, 10),
		Phase:                phase,
		Message:              message,
		ReconciledNamespaces: namespaces,
	}
	if streamRevision > 0 {
		status.StreamRevision = strconv.FormatUint(streamRevision, 10)
	}

	statusMap, err := toMap(status)
	if err != nil {
		return err
	}
	if err := unstructured.SetNestedMap(current.Object, statusMap, "status"); err != nil {
		return err
	}

	_, err = c.dynamic.Resource(clusterPolicyGVR).UpdateStatus(ctx, current, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("update policy status: %w", err)
	}
	return nil
}

func toMap(value any) (map[string]interface{}, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	out := map[string]interface{}{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}
