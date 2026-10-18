package validator

import (
	"testing"

	"devopsgen.dev/devopsgen/internal/output"
)

func TestValidate_DeploymentDuplicateContainerName(t *testing.T) {
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
        - name: app
          image: ghcr.io/x/app2:latest
`)},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("duplicate container name should be invalid: %v %v", r, err)
	}
}

func TestValidate_ServiceMissingTargetPort(t *testing.T) {
	arts := []output.Artifact{
		{Path: "kubernetes/service.yaml", Content: []byte(`
apiVersion: v1
kind: Service
metadata:
  name: app
spec:
  selector:
    app: app
  ports:
    - port: 8080
`)},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("missing targetPort should be invalid: %v %v", r, err)
	}
}

func TestValidate_WorkflowMissingOn(t *testing.T) {
	arts := []output.Artifact{
		{Path: ".github/workflows/ci.yml", Content: []byte(`
name: CI
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: docker push x:${{ github.sha }}
`)},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("missing on trigger should be invalid: %v %v", r, err)
	}
}

func TestValidate_WorkflowMissingJobs(t *testing.T) {
	arts := []output.Artifact{
		{Path: ".github/workflows/ci.yml", Content: []byte(`
name: CI
on:
  push:
`)},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("missing jobs should be invalid: %v %v", r, err)
	}
}

func TestValidate_WorkflowMissingName(t *testing.T) {
	arts := []output.Artifact{
		{Path: ".github/workflows/ci.yml", Content: []byte(`
on:
  push:
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: docker push x:${{ github.sha }}
`)},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("missing name should be invalid: %v %v", r, err)
	}
}

func TestValidate_DeploymentMissingSelector(t *testing.T) {
	arts := []output.Artifact{
		{Path: "kubernetes/deployment.yaml", Content: []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  template:
    metadata:
      labels:
        app: app
    spec:
      containers:
        - name: app
          image: ghcr.io/x/app:latest
`)},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("missing selector should be invalid: %v %v", r, err)
	}
}

func TestValidate_DockerfileShellSyntax(t *testing.T) {
	arts := []output.Artifact{
		{Path: "Dockerfile", Content: []byte("FROM scratch\nCOPY . . 2>/dev/null\n")},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("shell syntax in COPY should be invalid: %v %v", r, err)
	}
}

func TestValidate_DeploymentMissingTemplate(t *testing.T) {
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
`)},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("missing template should be invalid: %v %v", r, err)
	}
}
