package ipc

import (
	"context"

	"github.com/FilthyMudblood/aegis-fabric/internal/dataplane"
	afpsdk "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/sdk"
)

// PreflightService implements AFPSidecarIPC by delegating to SingleExecutionAuthority.
type PreflightService struct {
	afpsdk.UnimplementedAFPSidecarIPCServer
	sea *dataplane.SingleExecutionAuthority
}

func NewPreflightService(sea *dataplane.SingleExecutionAuthority) *PreflightService {
	return &PreflightService{sea: sea}
}

func (s *PreflightService) PreFlightCheck(_ context.Context, req *afpsdk.PreFlightRequest) (*afpsdk.PreFlightResponse, error) {
	if req == nil {
		req = &afpsdk.PreFlightRequest{}
	}
	return s.sea.EvaluatePreFlight(req.GetTraceId(), req.GetTargetDid(), req.GetEstimatedTasks()), nil
}

func (s *PreflightService) ReportInternalState(_ context.Context, req *afpsdk.InternalStateReport) (*afpsdk.StateAck, error) {
	if req == nil {
		req = &afpsdk.InternalStateReport{}
	}
	s.sea.ReportInternalState(req.GetCurrentRecursionDepth(), req.GetContextMemoryBytes())
	return &afpsdk.StateAck{Received: true}, nil
}
