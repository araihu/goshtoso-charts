package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionFlagReportsLinkerInjectedVersion(t *testing.T) {
	t.Parallel()

	binary := filepath.Join(t.TempDir(), "goshtoso-charts-server")
	const version = "v0.0.1-test"
	build := exec.Command("go", "build", "-ldflags", "-X github.com/araihu/goshtoso-charts/site/internal/buildinfo.siteVersion="+version, "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build injected version: %v\n%s", err, output)
	}

	output, err := exec.Command(binary, "-version").CombinedOutput()
	if err != nil {
		t.Fatalf("run version flag: %v\n%s", err, output)
	}
	if got := strings.TrimSpace(string(output)); got != version {
		t.Fatalf("version output = %q, want %q", got, version)
	}
}
