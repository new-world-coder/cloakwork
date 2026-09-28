// Package protocol is the length-prefixed JSON framing shared by Chrome native
// messaging and the local unix socket. Stdout of the native host is reserved
// for these frames; logs go to stderr.
package protocol

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

// MaxMessage is Chrome's native messaging limit.
const MaxMessage = 1024 * 1024

// Write sends one frame. The bytes are assembled first so a message is a
// single write on the pipe.
func Write(w io.Writer, payload []byte) error {
	if len(payload) == 0 || len(payload) > MaxMessage {
		return fmt.Errorf("invalid message length %d", len(payload))
	}
	frame := make([]byte, 4+len(payload))
	binary.LittleEndian.PutUint32(frame[:4], uint32(len(payload)))
	copy(frame[4:], payload)
	_, err := w.Write(frame)
	return err
}

// Read returns one JSON payload. A clean EOF before the header is io.EOF.
// A short body is an error.
func Read(r io.Reader) ([]byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, err
	}
	n := binary.LittleEndian.Uint32(hdr[:])
	if n == 0 || n > MaxMessage {
		return nil, fmt.Errorf("invalid message length %d", n)
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// Serve reads requests until the client closes the stream.
func Serve(r io.Reader, w io.Writer, handle func([]byte) []byte) error {
	for {
		msg, err := Read(r)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := Write(w, handle(msg)); err != nil {
			return err
		}
	}
}

// ErrorBody is a one-frame error response.
func ErrorBody(message string) []byte {
	b, err := json.Marshal(struct {
		Type  string `json:"type"`
		Error string `json:"error"`
	}{Type: "error", Error: message})
	if err != nil {
		return []byte(`{"type":"error","error":"invalid message"}`)
	}
	return b
}
