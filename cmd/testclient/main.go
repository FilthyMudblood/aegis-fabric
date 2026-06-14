package main

import (
	"encoding/binary"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/aegis-fabric/afp-sidecar/internal/dataplane"
	afpv1 "github.com/aegis-fabric/afp-sidecar/pkg/protocol/v1"
	"google.golang.org/protobuf/proto"
)

func main() {
	// 初始化基础设施级结构化日志
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// 连接到本地 Sidecar Ingress
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		slog.Error("failed to connect to Sidecar", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	slog.Info("connected to AFP Sidecar ingress")

	// 构建基准 GovernanceHeader
	header := &afpv1.GovernanceHeader{
		PacketId:        1,
		Version:         1,
		HysteresisEpoch: uint64(time.Now().Unix()),
		CoordinationTtl: 60,
		CvpScore:        0.95,
		EntropyLoad: &afpv1.EntropyLoad{
			ResourceAsymmetryRatio:   0.2,
			DependencyContentionRate: 0.1,
		},
		TopologyConsensusHash: []byte("e2e_test_signature_block"),
	}

	// ---------------------------------------------------------
	// 测试用例 1: 标准单帧发送 (Standard Frame)
	// ---------------------------------------------------------
	slog.Info("--- Test Case 1: Standard Frame ---")
	payload1, _ := proto.Marshal(header)
	if err := dataplane.WriteFrame(conn, payload1); err != nil {
		slog.Error("failed to write standard frame", "error", err)
	}
	time.Sleep(200 * time.Millisecond)

	// ---------------------------------------------------------
	// 测试用例 2: TCP 粘包模拟 (Sticky Packets)
	// 连续两个包在极短时间内拼接到同一个 Socket Buffer 写入
	// ---------------------------------------------------------
	slog.Info("--- Test Case 2: Sticky Packets (Multiple frames in one write) ---")
	header.PacketId = 2
	payload2, _ := proto.Marshal(header)
	header.PacketId = 3
	payload3, _ := proto.Marshal(header)

	buf := make([]byte, 0)
	lenBuf := make([]byte, 4)

	// 打包 Frame 2
	binary.BigEndian.PutUint32(lenBuf, uint32(len(payload2)))
	buf = append(buf, lenBuf...)
	buf = append(buf, payload2...)

	// 紧接打包 Frame 3
	binary.BigEndian.PutUint32(lenBuf, uint32(len(payload3)))
	buf = append(buf, lenBuf...)
	buf = append(buf, payload3...)

	// 物理层一次性推入
	if _, err := conn.Write(buf); err != nil {
		slog.Error("failed to write sticky packets", "error", err)
	}
	time.Sleep(200 * time.Millisecond)

	// ---------------------------------------------------------
	// 测试用例 3: TCP 半包截断模拟 (Partial Packets)
	// 故意将一个合法的帧切碎，分多次发送，测试 ReadFrame 的 I/O 阻塞机制
	// ---------------------------------------------------------
	slog.Info("--- Test Case 3: Partial Packets (Fragmented writes with delay) ---")
	header.PacketId = 4
	payload4, _ := proto.Marshal(header)

	fullFrame := make([]byte, 0)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(payload4)))
	fullFrame = append(fullFrame, lenBuf...)
	fullFrame = append(fullFrame, payload4...)

	// 发送前 5 个字节（4字节长度头 + 1字节内容）
	conn.Write(fullFrame[:5])
	slog.Info("sent partial frame (5 bytes), pausing to simulate network latency...")
	time.Sleep(500 * time.Millisecond) // 强制制造延迟

	// 发送剩余截断部分
	conn.Write(fullFrame[5:])
	slog.Info("sent remainder of the frame")

	time.Sleep(1 * time.Second)
	slog.Info("E2E tests completed successfully")
}
