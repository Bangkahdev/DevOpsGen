package generator

import (
	"testing"

	"devopsgen.dev/devopsgen/internal/config"
)

func TestSanitizeK8sName_DNS1035Label(t *testing.T) {
	cases := []struct{ in, want string }{
		{"my-app", "my-app"},
		{"my.app", "my-app"},     // service names forbid dots
		{"9lives", "app-9lives"}, // DNS-1035 forbids leading digit
		{"My-App", "my-app"},
		{"a..b", "a--b"},
		{"a.-b", "a--b"},
		{"", "app"},
		{"-_-", "app"},
	}
	for _, c := range cases {
		got := sanitizeK8sName(c.in)
		if got != c.want {
			t.Fatalf("sanitizeK8sName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	long := make([]byte, 0, 80)
	for len(long) < 80 {
		long = append(long, 'a')
	}
	if got := sanitizeK8sName(string(long)); len(got) > 63 {
		t.Fatalf("name longer than 63 chars: %d", len(got))
	}
}

func TestName_ValidForKubectl(t *testing.T) {
	cases := []string{"bangkah/my.app", "org/9lives", "x/a..b", "x/App-Name"}
	for _, img := range cases {
		cfg := config.Config{Image: img}
		name := templateData{Config: cfg}.Name()
		if name == "" {
			t.Fatalf("Name(%q) empty", img)
		}
		if name[0] < 'a' || name[0] > 'z' {
			t.Fatalf("Name(%q) = %q must start with a letter", img, name)
		}
		for i := 0; i < len(name); i++ {
			c := name[i]
			if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' {
				t.Fatalf("Name(%q) = %q contains invalid DNS-1035 char %q", img, name, c)
			}
		}
	}
}

func TestInstallCmdPerPackageManager(t *testing.T) {
	data := func(pm string) templateData {
		return templateData{Config: config.Config{PackageManager: pm, AppType: config.AppTypeGeneric, StartCmd: "npm start"}}
	}
	if got := data("pnpm").InstallCmd(); got != "npm install -g pnpm@10 && pnpm install --frozen-lockfile" {
		t.Fatalf("pnpm install: %q", got)
	}
	if got := data("yarn").InstallCmd(); got != "yarn install --frozen-lockfile" {
		t.Fatalf("yarn install: %q", got)
	}
	if got := data("npm").InstallCmd(); got != "if [ -f package-lock.json ]; then npm ci; else npm install; fi" {
		t.Fatalf("npm install: %q", got)
	}
	// Empty package manager defaults to npm.
	if got := data("").InstallCmd(); got != "if [ -f package-lock.json ]; then npm ci; else npm install; fi" {
		t.Fatalf("default install: %q", got)
	}

	if got := data("pnpm").PruneCmd(); got != "pnpm install --prod" {
		t.Fatalf("pnpm prune: %q", got)
	}
	if got := data("yarn").PruneCmd(); got != "yarn install --production" {
		t.Fatalf("yarn prune: %q", got)
	}
	if got := data("npm").PruneCmd(); got != "npm prune --omit=dev" {
		t.Fatalf("npm prune: %q", got)
	}
}

func TestGenerate_PnpmDockerfileUsesPnpm(t *testing.T) {
	cfg := config.Config{
		Stack:          config.StackNode,
		AppType:        config.AppTypeExpress,
		CI:             config.CIGitHubActions,
		Registry:       config.RegistryDockerHub,
		Image:          "org/api",
		Port:           3000,
		StartCmd:       "pnpm start",
		PackageManager: "pnpm",
	}
	arts, err := New().Generate(cfg)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	var df []byte
	for _, a := range arts {
		if a.Path == "Dockerfile" {
			df = a.Content
		}
	}
	if df == nil {
		t.Fatal("Dockerfile missing")
	}
	if !bytesContains(df, "npm install -g pnpm@10 && pnpm install --frozen-lockfile") {
		t.Fatal("Dockerfile must install pnpm before using it")
	}
	if !bytesContains(df, "pnpm start") {
		t.Fatal("Dockerfile must run pnpm start")
	}
}

func TestGenerate_Dockerignore(t *testing.T) {
	cfg := config.Config{
		Stack:    config.StackNode,
		AppType:  config.AppTypeExpress,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryDockerHub,
		Image:    "org/api",
		Port:     3000,
	}
	arts, err := New().Generate(cfg)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	var di []byte
	for _, a := range arts {
		if a.Path == ".dockerignore" {
			di = a.Content
		}
	}
	if di == nil {
		t.Fatal(".dockerignore not generated")
	}
	if !bytesContains(di, ".git") || !bytesContains(di, "node_modules") {
		t.Fatalf(".dockerignore missing required entries:\n%s", di)
	}
}

// bytesContains is a small helper to avoid importing strings everywhere.
func bytesContains(b []byte, sub string) bool {
	return len(sub) == 0 || indexOf(b, sub) >= 0
}

func indexOf(b []byte, sub string) int {
	n := len(sub)
	if n == 0 || n > len(b) {
		if n == 0 {
			return 0
		}
		return -1
	}
	for i := 0; i+n <= len(b); i++ {
		if string(b[i:i+n]) == sub {
			return i
		}
	}
	return -1
}
