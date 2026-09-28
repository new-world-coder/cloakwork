package install_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/cloakwork/cloakwork/daemon/internal/config"
	"github.com/cloakwork/cloakwork/daemon/internal/install"
)

func TestHostManifestMatchesTemplate(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	templatePath := filepath.Join(filepath.Dir(file), "..", "..", "..", "install", "com.cloakwork.daemon.json")
	body, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatal(err)
	}
	var template install.Manifest
	if err := json.Unmarshal(body, &template); err != nil {
		t.Fatal(err)
	}
	got := install.HostManifest("/ABS/PATH/TO/cloakwork-host")
	if template.Name != got.Name || template.Description != got.Description || template.Type != got.Type {
		t.Fatalf("template %#v generator %#v", template, got)
	}
	if len(template.AllowedOrigins) != 1 || template.AllowedOrigins[0] != config.ExtensionOrigin {
		t.Fatalf("origins %#v", template.AllowedOrigins)
	}
	if template.Path == "" || !strings.HasPrefix(template.Path, "/") {
		t.Fatalf("template path %q", template.Path)
	}
}

func TestInstallWritesManifestAndWrapper(t *testing.T) {
	root := t.TempDir()
	binary := filepath.Join(root, "src-cloakwork")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\necho cloakwork\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	support := filepath.Join(root, "support")
	hosts := filepath.Join(root, "NativeMessagingHosts")
	result, err := install.Install(install.Options{
		BinaryPath:     binary,
		SupportDir:     support,
		NativeHostsDir: hosts,
	})
	if err != nil {
		t.Fatal(err)
	}
	wrapper, err := os.ReadFile(result.WrapperPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(wrapper), " host\n") || !strings.Contains(string(wrapper), result.BinaryPath) {
		t.Fatalf("wrapper %s", wrapper)
	}
	manifestBody, err := os.ReadFile(result.ManifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest install.Manifest
	if err := json.Unmarshal(manifestBody, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Name != config.HostName || manifest.Type != "stdio" {
		t.Fatalf("manifest %#v", manifest)
	}
	if manifest.Path != result.WrapperPath {
		t.Fatalf("manifest path %s wrapper %s", manifest.Path, result.WrapperPath)
	}
	if manifest.AllowedOrigins[0] != "chrome-extension://EXTENSION_ID_PLACEHOLDER/" {
		t.Fatalf("origins %#v", manifest.AllowedOrigins)
	}
	if !filepath.IsAbs(manifest.Path) {
		t.Fatalf("path %s", manifest.Path)
	}
}
