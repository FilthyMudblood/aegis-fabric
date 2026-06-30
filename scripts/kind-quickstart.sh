#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CLUSTER_NAME="${KIND_CLUSTER_NAME:-afp}"
SIDECAR_IMAGE="ghcr.io/filthymudblood/aegis-fabric-sidecar:latest"
OPERATOR_IMAGE="ghcr.io/filthymudblood/aegis-fabric-operator:latest"
DEMO_AGENT_IMAGE="ghcr.io/filthymudblood/afp-demo-agent:latest"
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
docker build -t "${SIDECAR_IMAGE}" "${ROOT}"

echo "Building operator image..."
docker build -f "${ROOT}/Dockerfile.operator" -t "${OPERATOR_IMAGE}" "${ROOT}"

echo "Building demo agent image..."
docker build -f "${ROOT}/Dockerfile.demo-agent" -t "${DEMO_AGENT_IMAGE}" "${ROOT}"

echo "Loading images into kind..."
kind load docker-image "${SIDECAR_IMAGE}" --name "${CLUSTER_NAME}"
kind load docker-image "${OPERATOR_IMAGE}" --name "${CLUSTER_NAME}"
kind load docker-image "${DEMO_AGENT_IMAGE}" --name "${CLUSTER_NAME}"

echo "Applying AFP Kubernetes manifests..."
kubectl apply -f "${ROOT}/deploy/kubernetes/namespace.yaml"
kubectl apply -f "${ROOT}/deploy/kubernetes/configmap-afp.yaml"
kubectl apply -f "${ROOT}/deploy/kubernetes/crd/afpclusterpolicy.yaml"
kubectl apply -f "${ROOT}/deploy/kubernetes/examples/afpclusterpolicy-enterprise.yaml"
kubectl apply -f "${ROOT}/deploy/kubernetes/policy-controller-deployment.yaml"
kubectl apply -f "${ROOT}/deploy/kubernetes/operator-deployment.yaml"
kubectl apply -f "${ROOT}/deploy/kubernetes/agent-pod-demo.yaml"

echo "Waiting for agent pod..."
kubectl -n "${NAMESPACE}" wait --for=condition=Ready pod -l app.kubernetes.io/component=agent-node --timeout=180s

POD="$(kubectl -n "${NAMESPACE}" get pod -l app.kubernetes.io/component=agent-node -o jsonpath='{.items[0].metadata.name}')"

echo ""
echo "==> L2 recursion breaker (preflightclient, expect ISOLATED / exit 2)"
kubectl -n "${NAMESPACE}" exec "${POD}" -c afp-sidecar -- \
  preflightclient --recursion-depth 12 --estimated-tasks 1 || true

echo ""
echo "==> L2 intent burst (preflightclient, expect ISOLATED / exit 2)"
kubectl -n "${NAMESPACE}" exec "${POD}" -c afp-sidecar -- \
  preflightclient --recursion-depth 1 --estimated-tasks 10000 || true

echo ""
echo "==> L1 LangGraph planner (agent-core logs, expect annotated-stop)"
sleep 5
kubectl -n "${NAMESPACE}" logs "${POD}" -c agent-core --tail=20 || true

echo ""
echo "==> Patch cluster policy entropy limit to 0.80"
kubectl patch afpclusterpolicy enterprise-default --type merge -p '{"spec":{"entropyLimit":0.80}}' || true

cat <<EOF

Quickstart complete.

Watch the demo agent loop:
  kubectl -n ${NAMESPACE} logs -f ${POD} -c agent-core

Re-run policy operator / inspect hot-reload:
  go run ./cmd/operator
  kubectl -n ${NAMESPACE} get configmap afp-sidecar-config -o yaml
  kubectl -n ${NAMESPACE} exec ${POD} -c afp-sidecar -- ls -la /etc/afp/policy

EOF
