package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloakwork/cloakwork/daemon/internal/config"
	"github.com/cloakwork/cloakwork/daemon/internal/install"
	"github.com/cloakwork/cloakwork/daemon/internal/nativemsg"
	"github.com/cloakwork/cloakwork/daemon/internal/profile"
	"github.com/cloakwork/cloakwork/daemon/internal/protocol"
	"github.com/cloakwork/cloakwork/daemon/internal/service"
	"github.com/cloakwork/cloakwork/daemon/internal/sock"
)

func main() {
	if len(os.Args) < 2 {
		usage(os.Stderr)
		os.Exit(2)
	}
	code := 0
	switch os.Args[1] {
	case "status":
		code = cmdStatus()
	case "profile":
		code = cmdProfile(os.Args[2:])
	case "daemon":
		code = cmdDaemon()
	case "host":
		code = cmdHost()
	case "install-host":
		code = cmdInstall()
	case "help", "-h", "--help":
		usage(os.Stdout)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		usage(os.Stderr)
		code = 2
	}
	os.Exit(code)
}

func usage(w *os.File) {
	fmt.Fprintf(w, `CloakWork daemon %s

Usage:
  cloakwork status
  cloakwork profile [list|show [id]]
  cloakwork daemon
  cloakwork host
  cloakwork install-host

status exits 1 when the background daemon is not running.
host is started by Chrome over native messaging. It does not listen on a port.
`, config.Version)
}

func loadStore() (*profile.Store, int) {
	store, err := profile.LoadEmbedded()
	if err != nil {
		fmt.Fprintf(os.Stderr, "profile store: %v\n", err)
		return nil, 1
	}
	return store, 0
}

func cmdStatus() int {
	store, code := loadStore()
	if code != 0 {
		return code
	}
	path := config.SocketPath()
	conn, err := net.DialTimeout("unix", path, 200*time.Millisecond)
	if err != nil {
		fmt.Fprint(os.Stdout, service.FormatStatus(service.Status{
			Version:       config.Version,
			DaemonRunning: false,
			SocketPath:    path,
			ProfileID:     store.Active().ID,
			Host:          config.HostName,
		}))
		return 1
	}
	defer conn.Close()
	if err := protocol.Write(conn, []byte(`{"type":"getStatus"}`)); err != nil {
		fmt.Fprintf(os.Stderr, "status: %v\n", err)
		return 1
	}
	body, err := protocol.Read(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "status: %v\n", err)
		return 1
	}
	var resp struct {
		Type   string          `json:"type"`
		Error  string          `json:"error"`
		Status *service.Status `json:"status"`
	}
	if err := json.Unmarshal(body, &resp); err != nil || resp.Status == nil {
		fmt.Fprintf(os.Stderr, "status: unexpected response\n")
		return 1
	}
	fmt.Fprint(os.Stdout, service.FormatStatus(*resp.Status))
	if !resp.Status.DaemonRunning {
		return 1
	}
	return 0
}

func cmdProfile(args []string) int {
	store, code := loadStore()
	if code != 0 {
		return code
	}
	action := "show"
	id := ""
	if len(args) > 0 {
		action = args[0]
		if len(args) > 1 {
			id = args[1]
		}
	}
	switch action {
	case "list":
		for _, item := range store.IDs() {
			fmt.Fprintln(os.Stdout, item)
		}
		return 0
	case "show":
		p, err := store.Get(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
		body, err := json.MarshalIndent(p, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stdout, "%s\n", body)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown profile command %q\n", action)
		usage(os.Stderr)
		return 2
	}
}

func cmdDaemon() int {
	store, code := loadStore()
	if code != 0 {
		return code
	}
	path := config.SocketPath()
	svc := service.New(store, path, true)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	fmt.Fprintf(os.Stderr, "cloakwork daemon listening on %s\n", path)
	if err := (&sock.Server{Path: path, Handle: svc.Handle}).ListenAndServe(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "daemon: %v\n", err)
		return 1
	}
	return 0
}

func cmdHost() int {
	store, code := loadStore()
	if code != 0 {
		return code
	}
	svc := service.New(store, config.SocketPath(), false)
	if err := nativemsg.Host(os.Stdin, os.Stdout, svc.Handle); err != nil {
		fmt.Fprintf(os.Stderr, "host: %v\n", err)
		return 1
	}
	return 0
}

func cmdInstall() int {
	result, err := install.Install(install.Options{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "install-host: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stdout, "binary: %s\nwrapper: %s\nmanifest: %s\n", result.BinaryPath, result.WrapperPath, result.ManifestPath)
	fmt.Fprintf(os.Stdout, "allowed_origins: %s\n", config.ExtensionOrigin)
	return 0
}
