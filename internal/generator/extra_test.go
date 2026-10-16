package generator

import (
	"strings"
	"testing"

	"devopsgen.dev/devopsgen/internal/config"
)

func TestTemplateData_BuildOrDefault(t *testing.T) {
	cases := []struct {
		cfg    config.Config
		want   string
	}{
		{cfg: config.Config{Stack: config.StackGo, AppType: config.AppTypeBinary}, want: "go build -ldflags=\"-s -w\" -o /out/server ."},
		{cfg: config.Config{Stack: config.StackGo, AppType: config.AppTypeBinary, BuildCmd: "go build -o bin ."}, want: "go build -o bin ."},
		{cfg: config.Config{Stack: config.StackNode, AppType: config.AppTypeNextJS}, want: "npm run build"},
		{cfg: config.Config{Stack: config.StackNode, AppType: config.AppTypeNextJS, BuildCmd: "next build"}, want: "next build"},
		{cfg: config.Config{Stack: config.StackNode, AppType: config.AppTypeExpress}, want: ""},
		{cfg: config.Config{Stack: config.StackNode, AppType: config.AppTypeGeneric}, want: ""},
	}
	for i, c := range cases {
		td := templateData{Config: c.cfg}
		got := td.BuildOrDefault()
		if got != c.want {
			t.Fatalf("case %d: BuildOrDefault() = %q, want %q", i, got, c.want)
		}
	}
}

func TestTemplateData_StartOrDefault(t *testing.T) {
	cases := []struct {
		cfg    config.Config
		want   string
	}{
		{cfg: config.Config{Stack: config.StackGo, AppType: config.AppTypeBinary, Port: 8080}, want: "/server"},
		{cfg: config.Config{Stack: config.StackGo, AppType: config.AppTypeBinary, Port: 8080, StartCmd: "./mybin"}, want: "./mybin"},
		{cfg: config.Config{Stack: config.StackNode, AppType: config.AppTypeExpress, Port: 3000}, want: "node server.js"},
		{cfg: config.Config{Stack: config.StackNode, AppType: config.AppTypeNextJS, Port: 3000}, want: "npx next start -p 3000"},
		{cfg: config.Config{Stack: config.StackNode, AppType: config.AppTypeVite, Port: 8080}, want: "npx serve -s dist -l 8080"},
		{cfg: config.Config{Stack: config.StackNode, AppType: config.AppTypeNestJS, Port: 3000}, want: "node dist/main.js"},
		{cfg: config.Config{Stack: config.StackNode, AppType: config.AppTypeGeneric, Port: 3000}, want: "node server.js"},
	}
	for i, c := range cases {
		td := templateData{Config: c.cfg}
		got := td.StartOrDefault()
		if got != c.want {
			t.Fatalf("case %d: StartOrDefault() = %q, want %q", i, got, c.want)
		}
	}
}

func TestTemplateData_TestCmdForWorkflow(t *testing.T) {
	// Go: used as-is.
	td := templateData{Config: config.Config{Stack: config.StackGo, TestCmd: "go test ./..."}}
	if got := td.TestCmdForWorkflow(); got != "go test ./..." {
		t.Fatalf("Go test cmd: %q", got)
	}
	// Node.js: wrapped with npm run.
	td = templateData{Config: config.Config{Stack: config.StackNode, TestCmd: "jest"}}
	if got := td.TestCmdForWorkflow(); got != "npm run jest" {
		t.Fatalf("Node test cmd: %q", got)
	}
	// Empty: returns empty.
	td = templateData{Config: config.Config{Stack: config.StackGo}}
	if got := td.TestCmdForWorkflow(); got != "" {
		t.Fatalf("empty test cmd: %q", got)
	}
}

func TestTemplateData_Tags(t *testing.T) {
	td := templateData{Config: config.Config{Tag: "v1"}}
	tags := td.Tags()
	if len(tags) != 2 || tags[0] != "v1" || tags[1] != "${{ github.sha }}" {
		t.Fatalf("tags: %v", tags)
	}
	td = templateData{Config: config.Config{}}
	tags = td.Tags()
	if len(tags) != 2 || tags[0] != "latest" || tags[1] != "${{ github.sha }}" {
		t.Fatalf("default tags: %v", tags)
	}
}

func TestGenerate_GoUsesBuildCmd(t *testing.T) {
	cfg := config.Config{
		Stack:    config.StackGo,
		AppType:  config.AppTypeBinary,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryDockerHub,
		Image:    "a/b",
		Port:     8080,
		BuildCmd: "go build -o custom .",
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
		t.Fatal("Dockerfile not generated")
	}
	if !strings.Contains(string(df), "go build -o custom") {
		t.Fatalf("Dockerfile should use custom build cmd:\n%s", string(df))
	}
}
