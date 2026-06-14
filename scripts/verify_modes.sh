#!/bin/bash
set -e

echo "========================================================"
echo " AFP Sidecar Feature Gate Validation (E2E)"
echo "========================================================"

# 编译最新二进制
echo "[1/4] Compiling binaries..."
go build -o bin/sidecar ./cmd/sidecar
go build -o bin/modetester ./cmd/modetester

# ---------------------------------------------------------
# Test Case A: Enterprise Mesh Mode
# ---------------------------------------------------------
echo -e "\n[2/4] Testing Mode: ENTERPRISE-MESH (Core Only)"
export AFP_RUN_MODE="enterprise-mesh"

# 后台启动 Sidecar，将日志重定向以便分析
./bin/sidecar > mesh_test.log 2>&1 &
SIDECAR_PID=$!
sleep 2 # 等待端口绑定

echo " -> Injecting unauthorized stranger payload..."
./bin/modetester > /dev/null 2>&1

# 停止进程并分析日志
kill $SIDECAR_PID
wait $SIDECAR_PID 2>/dev/null || true

# 断言：内网模式下不征收陌生人税，应当放行 (Fast/Slow Path) 而非 Reject
if rg -q "stranger_tax_reject" mesh_test.log; then
	echo " ❌ FAILED: Enterprise mode should NOT enforce stranger tax!"
	exit 1
else
	echo " ✅ PASSED: Enterprise mode successfully bypassed stranger tax."
fi

# ---------------------------------------------------------
# Test Case B: Open Exchange Mode
# ---------------------------------------------------------
echo -e "\n[3/4] Testing Mode: OPEN-EXCHANGE (Core + Network)"
export AFP_RUN_MODE="open-exchange"

./bin/sidecar > open_test.log 2>&1 &
SIDECAR_PID=$!
sleep 2

echo " -> Injecting unauthorized stranger payload..."
./bin/modetester > /dev/null 2>&1

kill $SIDECAR_PID
wait $SIDECAR_PID 2>/dev/null || true

# 断言：公网模式下必须严格拦截无抵押物的陌生人
if rg -q "stranger tax validation failed" open_test.log; then
	echo " ✅ PASSED: Open Exchange mode successfully rejected stranger without collateral."
else
	echo " ❌ FAILED: Open Exchange mode failed to enforce stranger tax!"
	exit 1
fi

echo -e "\n[4/4] Cleaning"
unset AFP_RUN_MODE
echo "Done."
