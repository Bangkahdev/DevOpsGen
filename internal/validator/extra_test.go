package validator

import (
	"testing"

	"devopsgen.dev/devopsgen/internal/output"
)

func TestValidate_DeploymentMissingContainers(t *testing.T) {
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
      containers: []
`)},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("empty containers should be invalid: %v %v", r, err)
	}
}

func TestValidate_DeploymentMissingContainerName(t *testing.T) {
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
        - image: ghcr.io/x/app:latest
`)},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("missing container name should be invalid: %v %v", r, err)
	}
}

func TestValidate_ServiceMissingPort(t *testing.T) {
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
    - targetPort: 8080
`)},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("missing port should be invalid: %v %v", r, err)
	}
}

func TestValidate_IngressMissingHost(t *testing.T) {
	arts := []output.Artifact{
		{Path: "kubernetes/ingress.yaml", Content: []byte(`
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app
spec:
  rules:
    - http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: app
                port:
                  number: 8080
`)},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("missing host should be invalid: %v %v", r, err)
	}
}

func TestValidate_IngressMissingBackend(t *testing.T) {
	arts := []output.Artifact{
		{Path: "kubernetes/ingress.yaml", Content: []byte(`
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app
spec:
  rules:
    - host: app.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
`)},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("missing backend should be invalid: %v %v", r, err)
	}
}

func TestValidate_WorkflowMissingCheckout(t *testing.T) {
	arts := []output.Artifact{
		{Path: ".github/workflows/ci.yml", Content: []byte(`
name: CI
on:
  push:
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hello
      - run: docker push x:${{ github.sha }}
`)},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("missing checkout should be invalid: %v %v", r, err)
	}
}
