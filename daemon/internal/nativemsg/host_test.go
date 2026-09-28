package nativemsg_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/cloakwork/cloakwork/daemon/internal/nativemsg"
	"github.com/cloakwork/cloakwork/daemon/internal/profile"
	"github.com/cloakwork/cloakwork/daemon/internal/protocol"
	"github.com/cloakwork/cloakwork/daemon/internal/service"
)

func TestHostMessages(t *testing.T) {
	store, err := profile.LoadEmbedded()
	if err != nil {
		t.Fatal(err)
	}
	svc := service.New(store, "/tmp/cloakwork-test.sock", false)

	pong := roundTrip(t, svc, []byte(`{"type":"ping"}`))
	if pong["type"] != "pong" {
		t.Fatalf("ping response %#v", pong)
	}

	status := roundTrip(t, svc, []byte(`{"type":"getStatus"}`))
	if status["type"] != "status" {
		t.Fatalf("status response %#v", status)
	}
	body, _ := status["status"].(map[string]any)
	if body["profileId"] != profile.DefaultID {
		t.Fatalf("profileId = %#v", body["profileId"])
	}
	if body["daemonRunning"] != false {
		t.Fatalf("daemonRunning = %#v", body["daemonRunning"])
	}

	prof := roundTrip(t, svc, []byte(`{"type":"getProfile"}`))
	if prof["type"] != "profile" {
		t.Fatalf("profile response %#v", prof)
	}
	embedded, _ := prof["profile"].(map[string]any)
	if embedded["id"] != profile.DefaultID || embedded["platform"] != "Win32" {
		t.Fatalf("profile %#v", embedded)
	}

	unknown := roundTrip(t, svc, []byte(`{"type":"setTimezone"}`))
	if unknown["error"] != "unknown message type" {
		t.Fatalf("unknown %#v", unknown)
	}

	invalid := roundTrip(t, svc, []byte(`{"type":"ping","extra":1}`))
	if invalid["error"] != "invalid message" {
		t.Fatalf("invalid %#v", invalid)
	}
}

func roundTrip(t *testing.T, svc *service.Service, req []byte) map[string]any {
	t.Helper()
	var in bytes.Buffer
	if err := protocol.Write(&in, req); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := nativemsg.Host(&in, &out, svc.Handle); err != nil {
		t.Fatal(err)
	}
	body, err := protocol.Read(&out)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}
