package detector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDetect_Go(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "go.mod", "module example.com/app\n\ngo 1.22\n")
	r, err := New().Detect(dir)
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if r.Stack != "go" {
		t.Fatalf("stack: %s", r.Stack)
	}
	if !contains(r.Indicators, "go.mod") {
		t.Fatalf("indicators: %v", r.Indicators)
	}
	if !contains(r.AppCandidates, "binary") {
		t.Fatalf("app candidates: %v", r.AppCandidates)
	}
}

func TestDetect_NodeExpress(t *testing.T) {
	dir := t.TempDir()
	pkg := `{
  "name": "api",
  "scripts": { "build": "npm run build", "test": "jest", "start": "node server.js" },
  "dependencies": { "express": "^4" }
}`
	write(t, dir, "package.json", pkg)
	write(t, dir, "package-lock.json", "{}")
	r, err := New().Detect(dir)
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if r.Stack != "nodejs" {
		t.Fatalf("stack: %s", r.Stack)
	}
	if r.BuildCmd != "npm run build" {
		t.Fatalf("build cmd: %q", r.BuildCmd)
	}
	if r.TestCmd != "jest" {
		t.Fatalf("test cmd: %q", r.TestCmd)
	}
	if r.StartCmd != "node server.js" {
		t.Fatalf("start cmd: %q", r.StartCmd)
	}
	if r.PackageManager != "npm" {
		t.Fatalf("package manager: %q", r.PackageManager)
	}
	if !contains(r.AppCandidates, "express") {
		t.Fatalf("app candidates: %v", r.AppCandidates)
	}
}

func TestDetect_AmbiguousMultipleStacks(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "go.mod", "module x\n\ngo 1.22\n")
	write(t, dir, "package.json", `{"scripts":{"start":"node s.js"}}`)
	r, err := New().Detect(dir)
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if !r.IsAmbiguous() {
		t.Fatal("expected ambiguity")
	}
	if !strings.Contains(strings.Join(r.Ambiguities, "; "), "multiple stacks") {
		t.Fatalf("ambiguities: %v", r.Ambiguities)
	}
}

func TestDetect_NoIndicators(t *testing.T) {
	dir := t.TempDir()
	_, err := New().Detect(dir)
	if err == nil {
		t.Fatal("expected error for empty project")
	}
}

func TestDetect_NodeNoLockfile(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "package.json", `{"scripts":{"start":"node s.js"}}`)
	r, err := New().Detect(dir)
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if r.Lockfile != "" {
		t.Fatalf("expected no lockfile, got %q", r.Lockfile)
	}
	if !contains(r.Ambiguities, "no lockfile detected") {
		t.Fatalf("ambiguities should mention lockfile: %v", r.Ambiguities)
	}
}

func TestDetectionResult_HasHelpers(t *testing.T) {
	r := DetectionResult{Stack: "nodejs", StartCmd: "node s.js", BuildCmd: "npm run build", TestCmd: "jest"}
	if !r.HasStart() {
		t.Fatal("HasStart should be true for nodejs with start cmd")
	}
	if !r.HasBuild() {
		t.Fatal("HasBuild should be true")
	}
	if !r.HasTest() {
		t.Fatal("HasTest should be true")
	}

	empty := DetectionResult{Stack: "go"}
	if empty.HasStart() {
		t.Fatal("HasStart should be false for go")
	}
	if empty.HasBuild() {
		t.Fatal("HasBuild should be false when empty")
	}
	if empty.HasTest() {
		t.Fatal("HasTest should be false when empty")
	}
}

func TestDetect_NodeYarnLockfile(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "package.json", `{"scripts":{"start":"node s.js"},"dependencies":{"express":"^4"}}`)
	write(t, dir, "yarn.lock", "# yarn lockfile v1\n")
	r, err := New().Detect(dir)
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if r.Lockfile != "yarn.lock" {
		t.Fatalf("expected yarn.lock, got %q", r.Lockfile)
	}
	if r.PackageManager != "yarn" {
		t.Fatalf("expected yarn, got %q", r.PackageManager)
	}
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
