package dataplane

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
)

const (
	MaxFrameSize = 8 * 1024 * 1024 // 8MB 硬性防御上限，防止 OOM 攻击
)

var (
	ErrFrameTooLarge = errors.New("afp: frame size exceeds maximum allowed limit")
)

// ReadFrame strictly reads a Length-Value (LV) framed payload from the TCP stream.
// It prevents TCP stream fragmentation (粘包/半包) from corrupting the Protobuf decoder.
func ReadFrame(conn net.Conn) ([]byte, error) {
	// 1. Read 4-byte header for payload length (BigEndian)
	var length uint32
	err := binary.Read(conn, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}

	// 2. Validate payload length against physical limits
	if length > MaxFrameSize {
		return nil, ErrFrameTooLarge
	}

	// 3. Read exact payload bytes
	payload := make([]byte, length)
	_, err = io.ReadFull(conn, payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// WriteFrame encapsulates a payload with a 4-byte length prefix.
func WriteFrame(conn net.Conn, payload []byte) error {
	length := uint32(len(payload))
	err := binary.Write(conn, binary.BigEndian, length)
	if err != nil {
		return err
	}
	_, err = conn.Write(payload)
	return err
}
