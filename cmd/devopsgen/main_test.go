package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMainRuns(t *testing.T) {
	bin := buildBinary(t)
	out, err := exec.Command(bin).CombinedOutput()
	if err != nil {
		t.Fatalf("version command failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "devopsgen") {
		t.Fatalf("unexpected version output: %s", out)
	}
}

func TestMainHelp(t *testing.T) {
	bin := buildBinary(t)
	out, err := exec.Command(bin, "--help").CombinedOutput()
	if err != nil {
		t.Fatalf("help command failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Available Commands") {
		t.Fatalf("help missing commands: %s", out)
	}
}

func buildBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	name := "devopsgen-smoke"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	bin := filepath.Join(dir, name)
	cmd := exec.Command("go", "build", "-o", bin, ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	if _, err := os.Stat(bin); err != nil {
		t.Fatalf("binary not found after build: %v", err)
	}
	return bin
}