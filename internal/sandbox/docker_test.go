package sandbox

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func multiplexedFrame(streamType byte, payload string) []byte {
	frame := make([]byte, 8+len(payload))
	frame[0] = streamType
	binary.BigEndian.PutUint32(frame[4:8], uint32(len(payload)))
	copy(frame[8:], payload)
	return frame
}

func TestReadAttachedOutputSplitsStdoutAndStderr(t *testing.T) {
	stream := bytes.NewReader(append(
		append(multiplexedFrame(1, "hello\n"), multiplexedFrame(2, "bad news\n")...),
		multiplexedFrame(1, "world\n")...,
	))

	stdout, stderr, err := readAttachedOutput(stream)
	if err != nil {
		t.Fatalf("readAttachedOutput returned error: %v", err)
	}

	if stdout != "hello\nworld\n" {
		t.Fatalf("stdout mismatch: got %q", stdout)
	}

	if stderr != "bad news\n" {
		t.Fatalf("stderr mismatch: got %q", stderr)
	}
}
