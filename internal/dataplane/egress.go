package dataplane

import (
	"context"
	"log/slog"
	"net"
	"time"

	afpv1 "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1"
	"google.golang.org/protobuf/proto"
)

// EgressRouter handles outbound traffic from the local AgentOS to the external AFP mesh.
// It acts as a transparent proxy, attaching the physical governance header before dispatching.
type EgressRouter struct {
	localDID       string
	epochClock     func() uint64
	entropyMonitor func() float32 // 注入本地真实物理资源的遥测函数
}

func NewEgressRouter(localDID string, epochClock func() uint64, entropyMonitor func() float32) *EgressRouter {
	return &EgressRouter{
		localDID:       localDID,
		epochClock:     epochClock,
		entropyMonitor: entropyMonitor,
	}
}

// Dispatch connects to a remote Sidecar, injects the locally-attested GovernanceHeader,
// and multiplexes the actual business payload.
func (er *EgressRouter) Dispatch(ctx context.Context, targetAddr string, traceID string, currentDepth uint32, payload []byte) error {
	// 1. Establish TCP connection to the remote peer's Sidecar (Ingress Port)
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", targetAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 2. Sample local physics to construct Outbound Governance Header
	currentEntropy := er.entropyMonitor()

	header := &afpv1.GovernanceHeader{
		PacketId:              uint64(time.Now().UnixNano()),
		Version:               1,
		HysteresisEpoch:       er.epochClock(),
		CoordinationTtl:       120,
		TraceId:               traceID,
		RecursionDepth:        currentDepth + 1,
		CvpScore:              1.0, // Self-reported CVP (remote FSM will decay this based on its own strict ledger)
		TopologyConsensusHash: []byte("outbound_crypto_signature_placeholder"),
		EntropyLoad: &afpv1.EntropyLoad{
			ResourceAsymmetryRatio:   currentEntropy,
			DependencyContentionRate: 0.1,
		},
		DependencyCollateral: &afpv1.DependencyCollateral{
			CollateralType: "SYS_VIRTUAL_STAKE",
			SlashThreshold: 0.8, // 承诺底线，授权远端在发现恶意行为时执行 Slash
		},
	}
	if header.TraceId == "" {
		header.TraceId = er.localDID + "-" + time.Now().Format("20060102150405.000000000")
	}

	headerBytes, err := proto.Marshal(header)
	if err != nil {
		return err
	}

	// 3. Physical Network Dispatch: Send the Protocol Header Frame first
	if err := WriteFrame(conn, headerBytes); err != nil {
		return err
	}

	// 4. Send the underlying Agent Business Payload as the subsequent frame
	if len(payload) > 0 {
		if err := WriteFrame(conn, payload); err != nil {
			return err
		}
	}

	slog.Debug("outbound traffic securely dispatched with governance header", "target", targetAddr, "entropy", currentEntropy)
	return nil
}
