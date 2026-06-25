package main

import (
	"log/slog"
	"net"
	"os"

	"github.com/FilthyMudblood/aegis-fabric/internal/dataplane"
	afpv1 "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1"
	"google.golang.org/protobuf/proto"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	listenAddr := "127.0.0.1:8082"
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		slog.Error("Node Z failed to bind listener", "error", err)
		os.Exit(1)
	}
	defer listener.Close()

	slog.Info("Remote Node Z (Test Stub) started", "address", listenAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go func(c net.Conn) {
			defer c.Close()
			peerIP := c.RemoteAddr().String()

			// 验证读取带外治理头 (LV Frame)
			payload, err := dataplane.ReadFrame(c)
			if err != nil {
				slog.Error("Node Z failed to read frame", "error", err)
				return
			}

			var header afpv1.GovernanceHeader
			if err := proto.Unmarshal(payload, &header); err != nil {
				slog.Error("Node Z received malformed header", "error", err)
				return
			}

			// 验证是否携带了高昂的抵押物 (陌生人税)
			collateral := float32(0)
			if header.DependencyCollateral != nil {
				collateral = header.DependencyCollateral.SlashThreshold
			}

			slog.Info("Node Z received Stranger Tax connection",
				"source_ip", peerIP,
				"offered_cvp", header.CvpScore,
				"collateral_threshold", collateral)

		}(conn)
	}
}
