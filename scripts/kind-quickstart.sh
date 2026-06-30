#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CLUSTER_NAME="${KIND_CLUSTER_NAME:-afp}"
IMAGE="ghcr.io/filthymudblood/aegis-fabric-sidecar:latest"
NAMESPACE="afp-system"

require() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "error: missing required command: $1" >&2
    exit 1
  }
}

require kind
require kubectl
require docker

if ! kind get clusters | grep -qx "${CLUSTER_NAME}"; then
  echo "Creating kind cluster '${CLUSTER_NAME}'..."
  kind create cluster --name "${CLUSTER_NAME}"
fi

echo "Building sidecar image..."
docker build -t "${IMAGE}" "${ROOT}"

echo "Loading image into kind..."
kind load docker-image "${IMAGE}" --name "${CLUSTER_NAME}"

echo "Applying AFP Kubernetes manifests..."
kubectl apply -f "${ROOT}/deploy/kubernetes/namespace.yaml"
kubectl apply -f "${ROOT}/deploy/kubernetes/configmap-afp.yaml"
kubectl apply -f "${ROOT}/deploy/kubernetes/crd/afpclusterpolicy.yaml"
kubectl apply -f "${ROOT}/deploy/kubernetes/examples/afpclusterpolicy-enterprise.yaml"
kubectl apply -f "${ROOT}/deploy/kubernetes/agent-pod-template.yaml"

echo "Waiting for agent pod..."
kubectl -n "${NAMESPACE}" wait --for=condition=Ready pod -l app.kubernetes.io/component=agent-node --timeout=120s

POD="$(kubectl -n "${NAMESPACE}" get pod -l app.kubernetes.io/component=agent-node -o jsonpath='{.items[0].metadata.name}')"

echo ""
echo "==> Recursion breaker demo (expect ISOLATED / exit 2)"
kubectl -n "${NAMESPACE}" exec "${POD}" -c afp-sidecar -- \
  preflightclient --recursion-depth 12 --estimated-tasks 1 || true

echo ""
echo "==> Intent burst demo (expect ISOLATED / exit 2)"
kubectl -n "${NAMESPACE}" exec "${POD}" -c afp-sidecar -- \
  preflightclient --recursion-depth 1 --estimated-tasks 10000 || true

echo ""
echo "==> Patch cluster policy entropy limit to 0.80"
kubectl patch afpclusterpolicy enterprise-default --type merge -p '{"spec":{"entropyLimit":0.80}}' || true

cat <<EOF

Quickstart complete.

Next steps:
  1) Run the policy operator locally: go run ./cmd/operator
  2) Re-apply the patched AFPClusterPolicy and inspect ConfigMaps:
       kubectl -n ${NAMESPACE} get configmap afp-sidecar-config -o yaml
  3) Watch sidecar hot-reload directory inside the pod:
       kubectl -n ${NAMESPACE} exec ${POD} -c afp-sidecar -- ls -la /etc/afp/policy

EOF
