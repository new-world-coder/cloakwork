package protocol_test

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"
	"testing"

	"github.com/cloakwork/cloakwork/daemon/internal/protocol"
)

func TestFrameRoundTrip(t *testing.T) {
	payload := []byte(`{"type":"ping"}`)
	var buf bytes.Buffer
	if err := protocol.Write(&buf, payload); err != nil {
		t.Fatal(err)
	}
	got, err := protocol.Read(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("got %s", got)
	}
	if _, err := protocol.Read(&buf); err != io.EOF {
		t.Fatalf("expected EOF, got %v", err)
	}
}

func TestFrameRejectsOversizeLength(t *testing.T) {
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], protocol.MaxMessage+1)
	_, err := protocol.Read(bytes.NewReader(hdr[:]))
	if err == nil || !strings.Contains(err.Error(), "invalid message length") {
		t.Fatalf("err = %v", err)
	}
}

func TestFrameRejectsShortBody(t *testing.T) {
	var buf bytes.Buffer
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], 8)
	buf.Write(hdr[:])
	buf.Write([]byte("short"))
	_, err := protocol.Read(&buf)
	if err == nil {
		t.Fatal("expected short body error")
	}
}

func TestServeRoundTrip(t *testing.T) {
	var in bytes.Buffer
	if err := protocol.Write(&in, []byte(`{"type":"ping"}`)); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := protocol.Serve(&in, &out, func(req []byte) []byte {
		if string(req) != `{"type":"ping"}` {
			t.Fatalf("req %s", req)
		}
		return []byte(`{"type":"pong"}`)
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := protocol.Read(&out)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"type":"pong"}` {
		t.Fatalf("got %s", got)
	}
}
