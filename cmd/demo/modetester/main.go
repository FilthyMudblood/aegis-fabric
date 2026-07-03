package main

import (
	"encoding/binary"
	"log/slog"
	"net"
	"os"
	"time"

	afpv1 "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1"
	"google.golang.org/protobuf/proto"
)

func main() {
	// 构造一个绝对意义上的“陌生人恶意包”
	// 1. 未知 DID (触发陌生人税)
	// 2. 没有任何抵押物 (DependencyCollateral = nil)
	// 3. 携带伪造的高 CVP (试图骗过本地网关)
	header := &afpv1.GovernanceHeader{
		PacketId:              uint64(time.Now().UnixNano()),
		CvpScore:              0.99,
		HysteresisEpoch:       uint64(time.Now().Unix()),
		EntropyLoad:           &afpv1.EntropyLoad{ResourceAsymmetryRatio: 0.1},
		DependencyCollateral:  nil, // 致命缺失
		TopologyConsensusHash: []byte("modetester_signature"),
	}

	payload, _ := proto.Marshal(header)

	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		slog.Error("Modetester failed to connect to Sidecar ingress", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	// 写入 4 字节长度头
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(payload)))
	conn.Write(lenBuf)
	// 写入负载
	conn.Write(payload)

	slog.Info("Modetester dispatched unauthorized stranger payload", "target", "127.0.0.1:8080")
	time.Sleep(500 * time.Millisecond) // 等待 Sidecar 处理日志输出
}
