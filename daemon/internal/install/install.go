// Package install writes the Chrome native messaging host manifest for the
// current user. The extension never receives a socket it can open itself;
// Chrome starts the host listed in the manifest.
package install

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloakwork/cloakwork/daemon/internal/config"
)

const hostWrapperName = "cloakwork-host"

// Manifest is a Chrome native messaging host manifest.
type Manifest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Path           string   `json:"path"`
	Type           string   `json:"type"`
	AllowedOrigins []string `json:"allowed_origins"`
}

// HostManifest returns the manifest Chrome reads. hostPath must be absolute.
func HostManifest(hostPath string) Manifest {
	return Manifest{
		Name:           config.HostName,
		Description:    config.HostDescription,
		Path:           hostPath,
		Type:           "stdio",
		AllowedOrigins: []string{config.ExtensionOrigin},
	}
}

// Options selects where files are written. Empty dirs use the per-user Chrome
// and CloakWork directories on macOS.
type Options struct {
	BinaryPath     string
	SupportDir     string
	NativeHostsDir string
}

// Result is what Install wrote.
type Result struct {
	BinaryPath   string
	WrapperPath  string
	ManifestPath string
}

// Install copies the daemon binary, writes a stdio wrapper, and writes the
// Chrome host manifest. The wrapper is what allowed_origins is allowed to start.
func Install(opts Options) (Result, error) {
	binaryPath := opts.BinaryPath
	if binaryPath == "" {
		exe, err := os.Executable()
		if err != nil {
			return Result{}, err
		}
		binaryPath = exe
	}
	absBin, err := filepath.Abs(binaryPath)
	if err != nil {
		return Result{}, err
	}
	support := opts.SupportDir
	hosts := opts.NativeHostsDir
	if support == "" || hosts == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Result{}, err
		}
		if support == "" {
			support = filepath.Join(home, "Library", "Application Support", "CloakWork")
		}
		if hosts == "" {
			hosts = filepath.Join(home, "Library", "Application Support", "Google", "Chrome", "NativeMessagingHosts")
		}
	}
	if err := os.MkdirAll(support, 0o755); err != nil {
		return Result{}, err
	}
	if err := os.MkdirAll(hosts, 0o755); err != nil {
		return Result{}, err
	}

	destBin := filepath.Join(support, "cloakwork")
	if err := copyExecutable(absBin, destBin); err != nil {
		return Result{}, err
	}
	wrapper := filepath.Join(support, hostWrapperName)
	script := "#!/bin/sh\nexec " + shellQuote(destBin) + " host\n"
	if err := os.WriteFile(wrapper, []byte(script), 0o755); err != nil {
		return Result{}, err
	}

	manifestPath := filepath.Join(hosts, config.HostName+".json")
	body, err := json.MarshalIndent(HostManifest(wrapper), "", "  ")
	if err != nil {
		return Result{}, err
	}
	body = append(body, '\n')
	if err := os.WriteFile(manifestPath, body, 0o644); err != nil {
		return Result{}, err
	}
	return Result{BinaryPath: destBin, WrapperPath: wrapper, ManifestPath: manifestPath}, nil
}

func copyExecutable(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	inInfo, err := in.Stat()
	if err != nil {
		return err
	}
	if outInfo, err := os.Stat(dst); err == nil && os.SameFile(inInfo, outInfo) {
		return os.Chmod(dst, 0o755)
	}
	tmp := dst + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dst)
}

func shellQuote(path string) string {
	return "'" + strings.ReplaceAll(path, "'", `'\''`) + "'"
}
