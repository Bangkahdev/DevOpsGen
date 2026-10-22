package validator

import (
	"fmt"
	"testing"

	"devopsgen.dev/devopsgen/internal/output"
)

const auditfixDeployment = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
        - name: app
          image: ghcr.io/x/app:v1
          ports:
            - containerPort: 8080
              name: http
`

func auditfixRun(t *testing.T, files map[string]string) (Result, error) {
	t.Helper()
	arts := make([]output.Artifact, 0, len(files))
	for path, content := range files {
		arts = append(arts, output.Artifact{Path: path, Content: []byte(content)})
	}
	return New().Validate(arts)
}

func TestValidate_ParseYAMLMultiPropagatesBrokenSecondDoc(t *testing.T) {
	files := map[string]string{
		"kubernetes/deployment.yaml": auditfixDeployment + `---
this is: not: valid: yaml
`,
	}
	r, err := auditfixRun(t, files)
	if err == nil || r != ResultInvalid {
		t.Fatalf("broken second document must fail validation: %v %v", r, err)
	}
}

func TestValidate_SecondDocWrongShape(t *testing.T) {
	files := map[string]string{
		"kubernetes/deployment.yaml": auditfixDeployment + `---
just-a-scalar
`,
	}
	r, err := auditfixRun(t, files)
	if err == nil || r != ResultInvalid {
		t.Fatalf("scalar document must fail map decoding: %v %v", r, err)
	}
}

func TestValidate_PathTypeWhitelist(t *testing.T) {
	base := `apiVersion: v1
kind: Service
metadata:
  name: app
spec:
  selector:
    app: web
  ports:
    - port: 8080
      targetPort: 8080
---
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
            pathType: %s
            backend:
              service:
                name: app
                port:
                  number: 8080
`
	files := map[string]string{
		"kubernetes/service.yaml": fmt.Sprintf(base, "Bogus"),
	}
	r, err := auditfixRun(t, files)
	if err == nil || r != ResultInvalid {
		t.Fatalf("random pathType must be rejected: %v %v", r, err)
	}

	files = map[string]string{
		"kubernetes/service.yaml": fmt.Sprintf(base, "ImplementationSpecific"),
	}
	if r, err := auditfixRun(t, files); err != nil || r != ResultValid {
		t.Fatalf("fallback pathType must be accepted: %v %v", r, err)
	}
}

func TestValidate_ContainerPortRange(t *testing.T) {
	files := map[string]string{
		"kubernetes/deployment.yaml": `apiVersion: apps/v1
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
          image: ghcr.io/x/app:v1
          ports:
            - containerPort: 0
`,
	}
	r, err := auditfixRun(t, files)
	if err == nil || r != ResultInvalid {
		t.Fatalf("containerPort 0 must be rejected: %v %v", r, err)
	}

	files = map[string]string{
		"kubernetes/deployment.yaml": `apiVersion: apps/v1
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
          image: ghcr.io/x/app:v1
          ports:
            - containerPort: 65536
`,
	}
	r, err = auditfixRun(t, files)
	if err == nil || r != ResultInvalid {
		t.Fatalf("containerPort 65536 must be rejected: %v %v", r, err)
	}
}

func TestValidate_CrossNamesTargetPort(t *testing.T) {
	// Pod port is named "http"; a Service pointing at the unknown name "api"
	// must be flagged.
	svc := `apiVersion: v1
kind: Service
metadata:
  name: app
spec:
  selector:
    app: web
  ports:
    - port: 80
      targetPort: api
`
	files := map[string]string{
		"kubernetes/deployment.yaml": auditfixDeployment,
		"kubernetes/service.yaml":    svc,
	}
	r, err := auditfixRun(t, files)
	if err == nil || r != ResultInvalid {
		t.Fatalf("mismatched named targetPort must be rejected: %v %v", r, err)
	}

	svc = `apiVersion: v1
kind: Service
metadata:
  name: app
spec:
  selector:
    app: web
  ports:
    - port: 80
      targetPort: http
`
	files["kubernetes/service.yaml"] = svc
	r, err = auditfixRun(t, files)
	if err != nil || r != ResultValid {
		t.Fatalf("matching named targetPort should be valid: %v %v", r, err)
	}
}

func TestValidate_CrossServiceSelectorUnknownKey(t *testing.T) {
	svc := `apiVersion: v1
kind: Service
metadata:
  name: app
spec:
  selector:
    app: web
    extra: one
  ports:
    - port: 80
      targetPort: 8080
`
	files := map[string]string{
		"kubernetes/deployment.yaml": auditfixDeployment,
		"kubernetes/service.yaml":    svc,
	}
	r, err := auditfixRun(t, files)
	if err == nil || r != ResultInvalid {
		t.Fatalf("selector key missing from pod labels must be rejected: %v %v", r, err)
	}
}

func TestValidate_CrossIngressBackendPort(t *testing.T) {
	svc := `apiVersion: v1
kind: Service
metadata:
  name: app
spec:
  selector:
    app: web
  ports:
    - port: 80
      targetPort: 8080
`
	ingress := func(port string) string {
		return `apiVersion: networking.k8s.io/v1
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
            backend:
              service:
                name: app
                port:
                  number: ` + port + `
`
	}
	files := map[string]string{
		"kubernetes/service.yaml": svc,
		"kubernetes/ingress.yaml": ingress("3000"),
	}
	r, err := auditfixRun(t, files)
	if err == nil || r != ResultInvalid {
		t.Fatalf("ingress port not exposed by the service must be rejected: %v %v", r, err)
	}

	files["kubernetes/ingress.yaml"] = ingress("80")
	r, err = auditfixRun(t, files)
	if err != nil || r != ResultValid {
		t.Fatalf("ingress port matching service port should be valid: %v %v", r, err)
	}
}

func TestValidate_WorkflowInvalidYAML(t *testing.T) {
	workflow := `name: CI
on:
  push:
jobs:
  build-and-publish:
    runs-on: ubuntu-latest
    steps:
      - run: a: b
`
	files := map[string]string{
		".github/workflows/ci.yml": workflow,
	}
	r, err := auditfixRun(t, files)
	if err == nil || r != ResultInvalid {
		t.Fatalf("workflows must be parsed as YAML (FR-13): %v %v", r, err)
	}
}

func TestValidate_WorkflowTwoDocuments(t *testing.T) {
	files := map[string]string{
		".github/workflows/ci.yml": `name: CI
on:
  push:
jobs:
  build-and-publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: docker push x:${{ github.sha }}
---
name: Second
on:
  push:
jobs: {}
`,
	}
	r, err := auditfixRun(t, files)
	if err == nil || r != ResultInvalid {
		t.Fatalf("workflow files must contain exactly one document: %v %v", r, err)
	}
}

func TestValidate_WorkflowNoSteps(t *testing.T) {
	files := map[string]string{
		".github/workflows/ci.yml": `name: CI
on:
  push:
jobs:
  build-and-publish:
    runs-on: ubuntu-latest
    steps: []
`,
	}
	r, err := auditfixRun(t, files)
	if err == nil || r != ResultInvalid {
		t.Fatalf("job without steps must be rejected: %v %v", r, err)
	}
}
