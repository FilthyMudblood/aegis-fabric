package dataplane

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

func TestFrameRoundTrip(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	payload := []byte("hello-afp")
	done := make(chan error, 1)
	go func() {
		done <- WriteFrame(c1, payload)
	}()

	got, err := ReadFrame(c2)
	if err != nil {
		t.Fatalf("read frame failed: %v", err)
	}
	if string(got) != string(payload) {
		t.Fatalf("payload mismatch: got=%q want=%q", string(got), string(payload))
	}
	if err := <-done; err != nil {
		t.Fatalf("write frame failed: %v", err)
	}
}

func TestReadFrameTooLarge(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	go func() {
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, MaxFrameSize+1)
		_, _ = c1.Write(buf)
	}()

	_, err := ReadFrame(c2)
	if !errors.Is(err, ErrFrameTooLarge) {
		t.Fatalf("expected ErrFrameTooLarge, got %v", err)
	}
}

func TestReadFramePartialPayloadEOF(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c2.Close()

	go func() {
		defer c1.Close()
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, 8)
		_, _ = c1.Write(buf)
		_, _ = c1.Write([]byte("abc")) // send partial payload then close
	}()

	_, err := ReadFrame(c2)
	if err == nil {
		t.Fatal("expected error for partial payload, got nil")
	}
	if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected EOF/ErrUnexpectedEOF, got %v", err)
	}
}

func TestReadFrameZeroLength(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	go func() {
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, 0)
		_, _ = c1.Write(buf)
	}()

	_ = c2.SetReadDeadline(time.Now().Add(1 * time.Second))
	got, err := ReadFrame(c2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected zero-length payload, got len=%d", len(got))
	}
}
