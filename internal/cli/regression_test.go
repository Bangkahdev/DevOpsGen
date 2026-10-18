package cli

import (
	"strings"
	"testing"

	"devopsgen.dev/devopsgen/internal/config"
	"devopsgen.dev/devopsgen/internal/generator"
)

func goDockerfile(t *testing.T, appType config.AppType) string {
	t.Helper()
	arts, err := generator.New().Generate(config.Config{
		Stack:    config.StackGo,
		AppType:  appType,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryGHCR,
		Image:    "bangkah/my-app",
		Port:     8080,
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	for _, a := range arts {
		if a.Path == "Dockerfile" {
			return string(a.Content)
		}
	}
	t.Fatal("no Dockerfile artifact")
	return ""
}

// Regression: with --stack go and no --app-type, the build step rendered with
// no `go build` command at all, so the Dockerfile compiled nothing.
func TestGoDockerfile_EmptyAppTypeStillBuilds(t *testing.T) {
	for _, at := range []config.AppType{"", config.AppTypeBinary} {
		df := goDockerfile(t, at)
		found := false
		for _, line := range strings.Split(df, "\n") {
			if strings.HasPrefix(line, "RUN ") && strings.Contains(line, "go build") {
				found = true
			}
		}
		if !found {
			t.Errorf("app-type %q: no RUN instruction runs `go build`:\n%s", at, df)
		}
	}
}

// Regression: the runtime stage copied /src/devopsgen (the name of this very
// project) instead of the binary written by the build step.
func TestGoDockerfile_CopiesTheBinaryItBuilds(t *testing.T) {
	df := goDockerfile(t, "")
	if strings.Contains(df, "devopsgen") {
		t.Errorf("Dockerfile must not reference a hard-coded project name:\n%s", df)
	}
	if !strings.Contains(df, "-o /out/server") || !strings.Contains(df, "COPY --from=builder /out/server /server") {
		t.Errorf("build output path and COPY source must both be /out/server:\n%s", df)
	}
}

// A Dockerfile RUN must never end in bare variable assignments (shellcheck SC2034).
func TestGoDockerfile_NoRunEndsWithAssignment(t *testing.T) {
	df := goDockerfile(t, "")
	for _, line := range strings.Split(df, "\n") {
		if strings.HasPrefix(line, "RUN ") && strings.HasSuffix(strings.TrimSpace(line), "GOARCH=${TARGETARCH}") {
			t.Errorf("RUN line ends with an assignment and no command: %q", line)
		}
	}
}
