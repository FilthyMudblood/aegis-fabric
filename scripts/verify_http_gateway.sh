#!/bin/bash
set -e

echo "========================================================"
echo " AFP L7 Integration Tests (Against Running Containers) "
echo "========================================================"

# 等待容器就绪的辅助函数
wait_for_port() {
  local port=$1
  echo -n "Waiting for localhost:$port to be ready..."
  while ! nc -z localhost $port >/dev/null 2>&1; do
    sleep 1
    echo -n "."
  done
  echo " Ready!"
}

wait_for_port 8082
wait_for_port 8083

# ---------------------------------------------------------
# 场景 A: 企业内部网格 (Target: localhost:8082)
# ---------------------------------------------------------
echo -e "\n[1/2] Testing ENTERPRISE-MESH Node (Port 8082)..."

echo " -> Test 1: Normal Delegation (Expected: 200 OK)"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://127.0.0.1:8082/api/v1/agent/invoke \
  -H "X-AFP-Source-DID: did:afp:agent:alpha" \
  -H "X-AFP-Trace-ID: trace-normal-001" \
  -H "X-AFP-Recursion-Depth: 2" \
  -H "X-AFP-CVP: 0.9")

if [ "$HTTP_CODE" -eq 200 ]; then
    echo " ✅ PASSED: Normal request accepted."
else
    echo " ❌ FAILED: Expected 200, got $HTTP_CODE"; exit 1
fi

echo " -> Test 2: Infinite Loop Detection (Expected: 508 Loop Detected)"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://127.0.0.1:8082/api/v1/agent/invoke \
  -H "X-AFP-Source-DID: did:afp:agent:alpha" \
  -H "X-AFP-Trace-ID: trace-deadlock-002" \
  -H "X-AFP-Recursion-Depth: 11" \
  -H "X-AFP-CVP: 0.9")

if [ "$HTTP_CODE" -eq 508 ]; then
    echo " ✅ PASSED: Recursion breaker tripped correctly (508)."
else
    echo " ❌ FAILED: Expected 508, got $HTTP_CODE"; exit 1
fi

# ---------------------------------------------------------
# 场景 B: 开放交换局模式 (Target: localhost:8083)
# ---------------------------------------------------------
echo -e "\n[2/2] Testing OPEN-EXCHANGE Node (Port 8083)..."

echo " -> Test 3: Stranger Tax Enforcement (Expected: 403 Forbidden)"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://127.0.0.1:8083/api/v1/agent/invoke \
  -H "X-AFP-Source-DID: did:afp:hacker:unknown" \
  -H "X-AFP-Trace-ID: trace-attack-003" \
  -H "X-AFP-Recursion-Depth: 1" \
  -H "X-AFP-CVP: 0.1")

if [ "$HTTP_CODE" -eq 403 ]; then
    echo " ✅ PASSED: Stranger rejected by zero-trust gateway (403)."
else
    echo " ❌ FAILED: Expected 403, got $HTTP_CODE"; exit 1
fi

echo -e "\n🎉 All L7 HTTP blackbox integration tests passed."
