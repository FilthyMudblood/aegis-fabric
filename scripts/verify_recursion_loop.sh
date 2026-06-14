#!/bin/bash
set -euo pipefail

echo "========================================================"
echo " AFP Recursion Breaker Validation"
echo "========================================================"

mkdir -p bin
go build -o bin/sidecar ./cmd/sidecar
go build -o bin/egressclient ./cmd/egressclient
go build -o bin/looptester ./cmd/looptester

export AFP_METRICS_ADDR="127.0.0.1:19090"
./bin/sidecar > recursion_test.log 2>&1 &
SIDECAR_PID=$!
trap 'kill ${SIDECAR_PID} >/dev/null 2>&1 || true' EXIT
sleep 2

echo "[1/2] Dispatching normal recursion depth traffic..."
./bin/egressclient >/dev/null 2>&1 || true
sleep 1

echo "[2/2] Dispatching loop-depth payload (depth=10, expect reject at ingress)..."
./bin/looptester >/dev/null 2>&1 || true
sleep 1

kill ${SIDECAR_PID}
wait ${SIDECAR_PID} 2>/dev/null || true
trap - EXIT

if rg -q "recursion depth exceeded physical limit" recursion_test.log; then
  echo "✅ PASSED: recursion breaker rejected loop-depth request."
else
  echo "❌ FAILED: recursion breaker did not trigger."
  exit 1
fi

echo "Done. Inspect recursion_test.log for full trace."
