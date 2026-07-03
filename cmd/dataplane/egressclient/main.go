package main

import (
	"encoding/binary"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/FilthyMudblood/aegis-fabric/internal/dataplane"
)

func main() {
	// 初始化结构化日志
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// ---------------------------------------------------------
	// 测试用例: 递归寻址命中 (预期：本地 miss 后由核心邻居解析并回填缓存)
	// ---------------------------------------------------------
	slog.Info("--- Test Case: Recursive Topology Discovery (did:afp:remote:z) ---")
	testEgressFlow(
		"did:afp:remote:z",
		"Recursive discovery should resolve this DID and dispatch outbound traffic.",
		"trace-recursive-z-001",
		0,
	)

	time.Sleep(500 * time.Millisecond)
}

// testEgressFlow 模拟本地 AgentOS 使用升级后的 4 帧协议向 Sidecar 发送出站流量
func testEgressFlow(targetDID string, payload string, traceID string, currentDepth uint32) {
	// 连接到本地 Sidecar Egress 内部接口 (8081 端口)
	conn, err := net.Dial("tcp", "127.0.0.1:8081")
	if err != nil {
		slog.Error("failed to connect to local Sidecar egress", "error", err)
		return
	}
	defer conn.Close()

	slog.Info("connected to local egress", "target_did", targetDID)

	// 模拟升级后的本地 4 帧协议：
	// 第 1 帧: TraceID (用于全链路追踪)
	if err := dataplane.WriteFrame(conn, []byte(traceID)); err != nil {
		slog.Error("failed to write trace id frame", "error", err)
		return
	}

	// 第 2 帧: Current Depth (将其转为 4 字节大端序写入)
	depthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(depthBuf, currentDepth)
	if err := dataplane.WriteFrame(conn, depthBuf); err != nil {
		slog.Error("failed to write recursion depth frame", "error", err)
		return
	}

	// 第 3 帧: Target DID
	if err := dataplane.WriteFrame(conn, []byte(targetDID)); err != nil {
		slog.Error("failed to write target DID frame", "error", err)
		return
	}

	// 第 4 帧: Business Payload
	if err := dataplane.WriteFrame(conn, []byte(payload)); err != nil {
		slog.Error("failed to write payload frame", "error", err)
		return
	}

	slog.Info("local egress frames dispatched successfully")
}
