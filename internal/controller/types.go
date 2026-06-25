package controller

const (
	Group   = "afp.aegis-fabric.io"
	Version = "v1alpha1"
	Kind    = "AFPClusterPolicy"
	Resource = "afpclusterpolicies"

	SidecarConfigMapName = "afp-sidecar-config"
	AgentConfigMapName   = "afp-agent-config"
	DefaultNamespace     = "afp-system"
)

// ClusterPolicySpec mirrors the CRD spec and protobuf contract.
type ClusterPolicySpec struct {
	TargetNamespaces  []string `json:"targetNamespaces,omitempty"`
	EntropyLimit      float64  `json:"entropyLimit,omitempty"`
	MaxRecursionDepth uint32   `json:"maxRecursionDepth,omitempty"`
	RunMode           string   `json:"runMode,omitempty"`
	FailMode          string   `json:"failMode,omitempty"`
	MaxContextBytes   uint64   `json:"maxContextBytes,omitempty"`
}

// ClusterPolicy is the cluster-scoped AFP governance CRD.
type ClusterPolicy struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   map[string]any    `json:"metadata,omitempty"`
	Spec       ClusterPolicySpec `json:"spec,omitempty"`
}
