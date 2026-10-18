package validator

import (
	"testing"

	"devopsgen.dev/devopsgen/internal/output"
)

func TestCrossResource_ServiceSelectorMatchesDeployment(t *testing.T) {
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
          ports:
            - containerPort: 8080
`)},
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
      targetPort: 8080
`)},
	}
	r, err := New().Validate(arts)
	if err != nil || r != ResultValid {
		t.Fatalf("matching Service/Deployment should be valid: %v %v", r, err)
	}
}

func TestCrossResource_ServiceSelectorMismatchDeployment(t *testing.T) {
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
		{Path: "kubernetes/service.yaml", Content: []byte(`
apiVersion: v1
kind: Service
metadata:
  name: app
spec:
  selector:
    app: other
  ports:
    - port: 8080
      targetPort: 8080
`)},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("mismatched Service selector should be invalid: %v %v", r, err)
	}
}

func TestCrossResource_ServiceTargetPortMatchesDeployment(t *testing.T) {
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
          ports:
            - containerPort: 8080
`)},
		{Path: "kubernetes/service.yaml", Content: []byte(`
apiVersion: v1
kind: Service
metadata:
  name: app
spec:
  selector:
    app: app
  ports:
    - port: 80
      targetPort: 8080
`)},
	}
	r, err := New().Validate(arts)
	if err != nil || r != ResultValid {
		t.Fatalf("matching ports should be valid: %v %v", r, err)
	}
}

func TestCrossResource_ServiceTargetPortMismatchDeployment(t *testing.T) {
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
          ports:
            - containerPort: 8080
`)},
		{Path: "kubernetes/service.yaml", Content: []byte(`
apiVersion: v1
kind: Service
metadata:
  name: app
spec:
  selector:
    app: app
  ports:
    - port: 80
      targetPort: 9090
`)},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("mismatched targetPort should be invalid: %v %v", r, err)
	}
}

func TestCrossResource_IngressBackendMatchesService(t *testing.T) {
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
      targetPort: 8080
`)},
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
            backend:
              service:
                name: app
                port:
                  number: 8080
`)},
	}
	r, err := New().Validate(arts)
	if err != nil || r != ResultValid {
		t.Fatalf("matching Ingress backend should be valid: %v %v", r, err)
	}
}

func TestCrossResource_IngressBackendMismatchService(t *testing.T) {
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
      targetPort: 8080
`)},
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
            backend:
              service:
                name: other
                port:
                  number: 8080
`)},
	}
	r, err := New().Validate(arts)
	if err == nil || r != ResultInvalid {
		t.Fatalf("mismatched Ingress backend should be invalid: %v %v", r, err)
	}
}

func TestValidate_NoArtifacts(t *testing.T) {
	r, err := New().Validate([]output.Artifact{})
	if err == nil || r != ResultInvalid {
		t.Fatalf("empty artifacts should be invalid: %v %v", r, err)
	}
}
