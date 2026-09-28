// Package config holds process-wide constants shared by the CLI, the native
// messaging host, and the local socket server.
package config

import (
	"os"
	"path/filepath"
)

const (
	// Version is the Sprint 1 daemon version.
	Version = "0.1.0"
	// HostName is the Chrome native messaging host name.
	HostName = "com.cloakwork.daemon"
	// ExtensionOrigin is replaced with the real extension id when the Manifest
	// V3 extension exists. Chrome requires a trailing slash.
	ExtensionOrigin = "chrome-extension://EXTENSION_ID_PLACEHOLDER/"
	// HostDescription is the text Chrome shows for the native host.
	HostDescription = "CloakWork native messaging host"
)

// SocketPath is the per-user unix socket. On macOS, TempDir is the caller's
// private $TMPDIR, so the socket is not on a shared /tmp path.
func SocketPath() string {
	return filepath.Join(os.TempDir(), "cloakwork.sock")
}
