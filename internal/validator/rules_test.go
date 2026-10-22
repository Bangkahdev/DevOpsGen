package validator

import (
	"testing"

	"devopsgen.dev/devopsgen/internal/output"
)

func TestValidate_WorkflowAllowsBashDoubleBracket(t *testing.T) {
	arts := []output.Artifact{
		{
			Path: ".github/workflows/ci.yml",
			Content: []byte(`name: CI
on:
  push:
jobs:
  build-and-publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: |
          if [[ -f requirements.txt ]]; then pip install -r requirements.txt; fi
      - run: docker push x:${{ github.sha }}
`),
		},
	}
	r, err := New().Validate(arts)
	if err != nil || r != ResultValid {
		t.Fatalf("bash [[ -f ]] should be valid: %v %v", r, err)
	}
}

func TestValidate_WorkflowRejectsUnrenderedTemplate(t *testing.T) {
	arts := []output.Artifact{
		{
			Path: ".github/workflows/ci.yml",
			Content: []byte(`name: CI
on:
  push:
jobs:
  build-and-publish:
    runs-on: ubuntu-latest
    steps:
      - run: docker build -t {{.ImageRef}}:latest .
`),
		},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("unrendered {{. should be invalid: %v %v", r, err)
	}
}

func TestValidate_YAMLInvalidSyntax(t *testing.T) {
	arts := []output.Artifact{
		{Path: "kubernetes/deployment.yaml", Content: []byte("apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: app\nspec:\n  [invalid yaml")},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("invalid YAML should be rejected: %v %v", r, err)
	}
}

func TestValidate_DeploymentMultiDocument(t *testing.T) {
	// Deployment docs are fully validated; trailing docs of other kinds pass
	// through without deep validation (FR-13 still parses every document).
	arts := []output.Artifact{
		{Path: "kubernetes/deployment.yaml", Content: []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  selector:
    matchLabels:
      app: app
  template:
    metadata:
      labels:
        app: app
    spec:
      containers:
        - name: app
          image: ghcr.io/x/app:latest
---
apiVersion: v1
kind: Service
metadata:
  name: app
spec:
  selector:
    app: app
  ports:
    - port: 8080
      targetPort: 8080
`)},
	}
	r, err := New().Validate(arts)
	if err != nil || r != ResultValid {
		t.Fatalf("valid multi-doc: %v %v", r, err)
	}
}
