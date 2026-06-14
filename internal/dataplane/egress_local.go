package dataplane

import (
	"encoding/binary"
	"errors"
	"net"
)

const (
	MaxTargetDIDSize = 4 * 1024
	MaxTraceIDSize   = 8 * 1024
)

var (
	ErrInvalidTargetDID = errors.New("afp: invalid target did in local egress request")
	ErrInvalidTraceID   = errors.New("afp: invalid trace id in local egress request")
	ErrInvalidDepth     = errors.New("afp: invalid recursion depth frame")
)

// ReadLocalEgressRequest reads a four-frame local egress request:
// frame-1 trace ID, frame-2 recursion depth (4-byte BE), frame-3 target DID, frame-4 payload.
func ReadLocalEgressRequest(conn net.Conn) (string, uint32, string, []byte, error) {
	traceBytes, err := ReadFrame(conn)
	if err != nil {
		return "", 0, "", nil, err
	}
	if len(traceBytes) == 0 || len(traceBytes) > MaxTraceIDSize {
		return "", 0, "", nil, ErrInvalidTraceID
	}

	depthBuf, err := ReadFrame(conn)
	if err != nil {
		return "", 0, "", nil, err
	}
	if len(depthBuf) != 4 {
		return "", 0, "", nil, ErrInvalidDepth
	}
	currentDepth := binary.BigEndian.Uint32(depthBuf)

	targetBytes, err := ReadFrame(conn)
	if err != nil {
		return "", 0, "", nil, err
	}
	if len(targetBytes) == 0 || len(targetBytes) > MaxTargetDIDSize {
		return "", 0, "", nil, ErrInvalidTargetDID
	}

	payload, err := ReadFrame(conn)
	if err != nil {
		return "", 0, "", nil, err
	}
	return string(traceBytes), currentDepth, string(targetBytes), payload, nil
}
