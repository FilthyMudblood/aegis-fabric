#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${1:-${ROOT}/.afp-tls}"
NAMESPACE="${AFP_NAMESPACE:-afp-system}"

mkdir -p "${OUT_DIR}"
go run "${ROOT}/cmd/gencerts" --out "${OUT_DIR}"

kubectl create secret generic afp-policy-tls \
  --namespace "${NAMESPACE}" \
  --from-file=ca.crt="${OUT_DIR}/ca.crt" \
  --from-file=server.crt="${OUT_DIR}/server.crt" \
  --from-file=server.key="${OUT_DIR}/server.key" \
  --from-file=client.crt="${OUT_DIR}/client.crt" \
  --from-file=client.key="${OUT_DIR}/client.key" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "applied secret/afp-policy-tls in ${NAMESPACE}"
