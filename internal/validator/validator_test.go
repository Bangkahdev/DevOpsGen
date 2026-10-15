package validator

import (
	"testing"

	"devopsgen.dev/devopsgen/internal/output"
)

func TestValidate_Dockerfile(t *testing.T) {
	arts := []output.Artifact{
		{Path: "Dockerfile", Content: []byte("FROM golang:1.22\nEXPOSE 8080")},
	}
	r, err := New().Validate(arts)
	if err != nil || r != ResultValid {
		t.Fatalf("valid: %v %v", r, err)
	}
}

func TestValidate_DockerfileMissingFrom(t *testing.T) {
	arts := []output.Artifact{
		{Path: "Dockerfile", Content: []byte("EXPOSE 8080")},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("invalid: %v %v", r, err)
	}
}

func TestValidate_DockerfileCredentialLogin(t *testing.T) {
	arts := []output.Artifact{
		{Path: "Dockerfile", Content: []byte("FROM x\nRUN docker login -u foo")},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("should reject credential login: %v %v", r, err)
	}
}

func TestValidate_Deployment(t *testing.T) {
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
`)},
	}
	r, err := New().Validate(arts)
	if err != nil || r != ResultValid {
		t.Fatalf("valid: %v %v", r, err)
	}
}

func TestValidate_DeploymentMissingFields(t *testing.T) {
	arts := []output.Artifact{
		{Path: "kubernetes/deployment.yaml", Content: []byte("kind: Deployment\n")},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("invalid: %v %v", r, err)
	}
}

func TestValidate_Workflow(t *testing.T) {
	arts := []output.Artifact{
		{Path: ".github/workflows/ci.yml", Content: []byte(`
name: CI
on:
  push:
jobs:
  build-and-publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: docker push x:${{ github.sha }}
`)},
	}
	r, err := New().Validate(arts)
	if err != nil || r != ResultValid {
		t.Fatalf("valid: %v %v", r, err)
	}
}

func TestValidate_WorkflowMissingSHA(t *testing.T) {
	arts := []output.Artifact{
		{Path: ".github/workflows/ci.yml", Content: []byte("name: CI\non:\njobs:\n  x:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n")},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("should reject workflow without SHA tag: %v %v", r, err)
	}
}

func TestValidate_Empty(t *testing.T) {
	r, err := New().Validate(nil)
	if err == nil || r != ResultInvalid {
		t.Fatalf("empty should be invalid: %v %v", r, err)
	}
}

func TestValidate_Service(t *testing.T) {
	arts := []output.Artifact{
		{Path: "kubernetes/service.yaml", Content: []byte("kind: Service\nspec:\n  ports:\n    - targetPort: 8080\n")},
	}
	r, err := New().Validate(arts)
	if err != nil || r != ResultValid {
		t.Fatalf("valid: %v %v", r, err)
	}
}

func TestValidate_Ingress(t *testing.T) {
	arts := []output.Artifact{
		{Path: "kubernetes/ingress.yaml", Content: []byte("kind: Ingress\napiVersion: networking.k8s.io/v1\nspec:\n  rules:\n    - http:\n        paths:\n          - pathType: Prefix\n")},
	}
	r, err := New().Validate(arts)
	if err != nil || r != ResultValid {
		t.Fatalf("valid: %v %v", r, err)
	}
}

func TestValidate_IngressMissingAPI(t *testing.T) {
	arts := []output.Artifact{
		{Path: "kubernetes/ingress.yaml", Content: []byte("kind: Ingress\n")},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("should reject ingress without apiVersion: %v %v", r, err)
	}
}

func TestValidate_UnknownArtifactSkipped(t *testing.T) {
	arts := []output.Artifact{
		{Path: "README.md", Content: []byte("# hi")},
	}
	r, err := New().Validate(arts)
	if err != nil || r != ResultValid {
		t.Fatalf("unknown artifacts should pass: %v %v", r, err)
	}
}

func TestValidateResultStrings(t *testing.T) {
	if ResultValid != 0 || ResultInvalid != 1 || ResultExternalSkipped != 2 {
		t.Fatal("result constants mismatch")
	}
	if ResultValid.String() != "valid" {
		t.Fatalf("expected valid, got %q", ResultValid.String())
	}
	if ResultInvalid.String() != "invalid" {
		t.Fatalf("expected invalid, got %q", ResultInvalid.String())
	}
	if ResultExternalSkipped.String() != "external_skipped" {
		t.Fatalf("expected external_skipped, got %q", ResultExternalSkipped.String())
	}
	if Result(99).String() != "unknown" {
		t.Fatalf("expected unknown, got %q", Result(99).String())
	}
}
