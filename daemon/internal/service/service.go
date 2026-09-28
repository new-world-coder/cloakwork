// Package service answers ping, getStatus, and getProfile for both transports.
package service

import (
	"bytes"
	"encoding/json"
	"net"
	"time"

	"github.com/cloakwork/cloakwork/daemon/internal/config"
	"github.com/cloakwork/cloakwork/daemon/internal/profile"
	"github.com/cloakwork/cloakwork/daemon/internal/protocol"
)

// Status is the getStatus payload.
type Status struct {
	Version       string `json:"version"`
	DaemonRunning bool   `json:"daemonRunning"`
	SocketPath    string `json:"socketPath"`
	ProfileID     string `json:"profileId"`
	Host          string `json:"host"`
}

// Service is the shared request handler. Local is true when this process is
// the daemon, so getStatus does not dial its own socket.
type Service struct {
	Store      *profile.Store
	SocketPath string
	Local      bool
}

// New returns a handler over the embedded cohorts.
func New(store *profile.Store, socketPath string, local bool) *Service {
	return &Service{Store: store, SocketPath: socketPath, Local: local}
}

type request struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
}

type response struct {
	Type    string           `json:"type"`
	Error   string           `json:"error,omitempty"`
	Profile *profile.Profile `json:"profile,omitempty"`
	Status  *Status          `json:"status,omitempty"`
}

// Handle decodes one request and returns a JSON response body.
func (s *Service) Handle(body []byte) []byte {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	var req request
	if err := dec.Decode(&req); err != nil || dec.More() {
		return protocol.ErrorBody("invalid message")
	}
	switch req.Type {
	case "ping":
		return mustJSON(response{Type: "pong"})
	case "getStatus":
		st := s.status()
		return mustJSON(response{Type: "status", Status: &st})
	case "getProfile":
		p, err := s.Store.Get(req.ID)
		if err != nil {
			return protocol.ErrorBody("unknown profile")
		}
		return mustJSON(response{Type: "profile", Profile: &p})
	case "":
		return protocol.ErrorBody("invalid message")
	default:
		return protocol.ErrorBody("unknown message type")
	}
}

func (s *Service) status() Status {
	return Status{
		Version:       config.Version,
		DaemonRunning: s.daemonRunning(),
		SocketPath:    s.SocketPath,
		ProfileID:     s.Store.Active().ID,
		Host:          config.HostName,
	}
}

func (s *Service) daemonRunning() bool {
	if s.Local {
		return true
	}
	conn, err := net.DialTimeout("unix", s.SocketPath, 200*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// FormatStatus renders a status block for the CLI.
func FormatStatus(st Status) string {
	state := "stopped"
	if st.DaemonRunning {
		state = "running"
	}
	return "daemon: " + state + "\n" +
		"socket: " + st.SocketPath + "\n" +
		"profile: " + st.ProfileID + "\n" +
		"host: " + st.Host + "\n" +
		"version: " + st.Version + "\n"
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return protocol.ErrorBody("invalid message")
	}
	return b
}
