package policyplane

import (
	"context"
	"fmt"
	"time"

	afppolicystream "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1/policystream"
	"github.com/google/uuid"
)

// Server implements AFPPolicyStream for runtime policy push.
type Server struct {
	afppolicystream.UnimplementedAFPPolicyStreamServer
	hub *Hub
}

func NewServer(hub *Hub) *Server {
	return &Server{hub: hub}
}

func (s *Server) StreamPolicyUpdates(
	req *afppolicystream.PolicySubscribeRequest,
	stream afppolicystream.AFPPolicyStream_StreamPolicyUpdatesServer,
) error {
	if req == nil {
		return fmt.Errorf("policyplane: subscribe request is required")
	}
	sidecarID := req.GetSidecarId()
	if sidecarID == "" {
		return fmt.Errorf("policyplane: sidecar_id is required")
	}

	updates, unsubscribe := s.hub.Subscribe(sidecarID, req.GetLastRevision())
	defer unsubscribe()

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			if update == nil {
				continue
			}
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
}

func (s *Server) PublishPolicyUpdate(
	ctx context.Context,
	update *afppolicystream.PolicyUpdate,
) (*afppolicystream.PolicyUpdateAck, error) {
	_ = ctx
	if update == nil {
		return &afppolicystream.PolicyUpdateAck{
			Accepted: false,
			Message:  "update payload is required",
		}, nil
	}

	if update.GetUpdateId() == "" {
		update.UpdateId = uuid.NewString()
	}
	if update.GetIssuedAtUnix() == 0 {
		update.IssuedAtUnix = time.Now().Unix()
	}
	if update.GetSource() == afppolicystream.PolicySource_POLICY_SOURCE_UNSPECIFIED {
		update.Source = afppolicystream.PolicySource_OPERATOR_CLI
	}

	revision := s.hub.Publish(update)
	return &afppolicystream.PolicyUpdateAck{
		Accepted: true,
		Message:  "policy update published",
		Revision: revision,
	}, nil
}
