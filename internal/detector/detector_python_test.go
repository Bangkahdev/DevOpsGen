package detector

import (
	"testing"
)

func TestDetect_Python(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "pyproject.toml", "[project]\nname = 'api'\ndependencies = ['fastapi']\n")
	write(t, dir, "requirements.txt", "fastapi\n")
	r, err := New().Detect(dir)
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if r.Stack != "python" {
		t.Fatalf("stack: %s", r.Stack)
	}
	if !contains(r.Indicators, "pyproject.toml") && !contains(r.Indicators, "requirements.txt") {
		t.Fatalf("indicators: %v", r.Indicators)
	}
	if !contains(r.AppCandidates, "fastapi") {
		t.Fatalf("app candidates: %v", r.AppCandidates)
	}
}

func TestDetect_PythonDjango(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "manage.py", "print('ok')")
	write(t, dir, "requirements.txt", "django\n")
	r, err := New().Detect(dir)
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if !contains(r.AppCandidates, "django") {
		t.Fatalf("app candidates: %v", r.AppCandidates)
	}
}

func TestDetect_PythonFlask(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "requirements.txt", "flask\n")
	r, err := New().Detect(dir)
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if !contains(r.AppCandidates, "flask") {
		t.Fatalf("app candidates: %v", r.AppCandidates)
	}
}

func TestDetect_PythonNoIndicators(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "README.md", "# hi")
	_, err := New().Detect(dir)
	if err == nil {
		t.Fatal("expected error for non-python project")
	}
}

func TestDetect_PythonAndGoAmbiguous(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "go.mod", "module x\n\ngo 1.22\n")
	write(t, dir, "pyproject.toml", "[project]\nname='x'\n")
	r, err := New().Detect(dir)
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if !r.IsAmbiguous() {
		t.Fatal("expected ambiguity for go+python")
	}
}

func TestDetect_PythonLockfiles(t *testing.T) {
	// poetry.lock
	dir1 := t.TempDir()
	write(t, dir1, "poetry.lock", "")
	write(t, dir1, "pyproject.toml", "")
	r1, err := New().Detect(dir1)
	if err != nil || r1.PackageManager != "poetry" || r1.Lockfile != "poetry.lock" {
		t.Fatalf("expected poetry, got %v %v", r1, err)
	}

	// Pipfile.lock
	dir2 := t.TempDir()
	write(t, dir2, "Pipfile", "")
	write(t, dir2, "Pipfile.lock", "")
	r2, err := New().Detect(dir2)
	if err != nil || r2.PackageManager != "pipenv" || r2.Lockfile != "Pipfile.lock" {
		t.Fatalf("expected pipenv, got %v %v", r2, err)
	}

	// uv.lock
	dir3 := t.TempDir()
	write(t, dir3, "uv.lock", "")
	write(t, dir3, "pyproject.toml", "")
	r3, err := New().Detect(dir3)
	if err != nil || r3.PackageManager != "uv" || r3.Lockfile != "uv.lock" {
		t.Fatalf("expected uv, got %v %v", r3, err)
	}
}
