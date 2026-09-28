// Package sock is the per-user unix socket used by the CLI and, later, the SDK.
// On macOS the peer uid comes from LOCAL_PEERCRED. Connections from another
// uid are rejected. There is no TCP listener.
package sock

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/cloakwork/cloakwork/daemon/internal/protocol"
	"golang.org/x/sys/unix"
)

// PeerFunc supplies the peer uid. Tests use it to simulate another user.
// A nil func uses the OS credential lookup.
type PeerFunc func(*net.UnixConn) (int, error)

// Server accepts framed JSON connections on a unix socket.
type Server struct {
	Path   string
	Handle func([]byte) []byte
	Peer   PeerFunc
}

// ListenAndServe owns Path until ctx is cancelled.
func (s *Server) ListenAndServe(ctx context.Context) error {
	if conn, err := net.Dial("unix", s.Path); err == nil {
		conn.Close()
		return fmt.Errorf("daemon already running on %s", s.Path)
	}
	if err := os.Remove(s.Path); err != nil && !os.IsNotExist(err) {
		return err
	}

	old := unix.Umask(0o077)
	ln, err := net.Listen("unix", s.Path)
	unix.Umask(old)
	if err != nil {
		return err
	}
	if err := os.Chmod(s.Path, 0o600); err != nil {
		ln.Close()
		return err
	}

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				os.Remove(s.Path)
				return nil
			}
			return err
		}
		uc, ok := conn.(*net.UnixConn)
		if !ok {
			conn.Close()
			continue
		}
		go s.serve(uc)
	}
}

func (s *Server) serve(conn *net.UnixConn) {
	defer conn.Close()
	if !s.allow(conn) {
		_ = protocol.Write(conn, protocol.ErrorBody("unauthorized"))
		return
	}
	for {
		msg, err := protocol.Read(conn)
		if err == io.EOF {
			return
		}
		if err != nil {
			return
		}
		if err := protocol.Write(conn, s.Handle(msg)); err != nil {
			return
		}
	}
}

func (s *Server) allow(conn *net.UnixConn) bool {
	lookup := s.Peer
	if lookup == nil {
		lookup = peerUID
	}
	uid, err := lookup(conn)
	if err != nil {
		return false
	}
	return uid == os.Getuid()
}
