package sock_test

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/cloakwork/cloakwork/daemon/internal/profile"
	"github.com/cloakwork/cloakwork/daemon/internal/protocol"
	"github.com/cloakwork/cloakwork/daemon/internal/service"
	"github.com/cloakwork/cloakwork/daemon/internal/sock"
)

func TestSocketRejectsDifferentUID(t *testing.T) {
	path, stop := startServer(t, func(*net.UnixConn) (int, error) {
		return os.Getuid() + 1, nil
	})
	defer stop()

	conn, err := net.Dial("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := protocol.Write(conn, []byte(`{"type":"ping"}`)); err != nil {
		t.Fatal(err)
	}
	body, err := protocol.Read(conn)
	if err != nil {
		t.Fatal(err)
	}
	var resp map[string]string
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp["type"] != "error" || resp["error"] != "unauthorized" {
		t.Fatalf("response %#v", resp)
	}
}

func TestSocketAcceptsSameUser(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("production peer check uses getpeereid on macOS")
	}
	path, stop := startServer(t, nil)
	defer stop()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("socket mode %o", info.Mode().Perm())
	}

	conn, err := net.Dial("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := protocol.Write(conn, []byte(`{"type":"getProfile"}`)); err != nil {
		t.Fatal(err)
	}
	body, err := protocol.Read(conn)
	if err != nil {
		t.Fatal(err)
	}
	var resp struct {
		Type    string `json:"type"`
		Error   string `json:"error"`
		Profile struct {
			ID string `json:"id"`
		} `json:"profile"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error != "" || resp.Type != "profile" || resp.Profile.ID != profile.DefaultID {
		t.Fatalf("response %#v", resp)
	}
}

func startServer(t *testing.T, peer sock.PeerFunc) (string, func()) {
	t.Helper()
	// macOS unix socket paths must fit in sockaddr_un (104 bytes). t.TempDir
	// under /var/folders is too long, so the socket lives in /tmp.
	dir, err := os.MkdirTemp("/tmp", "cw")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	path := filepath.Join(dir, "s.sock")
	store, err := profile.LoadEmbedded()
	if err != nil {
		t.Fatal(err)
	}
	svc := service.New(store, path, true)
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- (&sock.Server{Path: path, Handle: svc.Handle, Peer: peer}).ListenAndServe(ctx)
	}()
	deadline := time.Now().Add(2 * time.Second)
	for {
		conn, err := net.Dial("unix", path)
		if err == nil {
			conn.Close()
			break
		}
		if time.Now().After(deadline) {
			cancel()
			t.Fatalf("server did not listen: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	return path, func() {
		cancel()
		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("server: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("server did not stop")
		}
	}
}
