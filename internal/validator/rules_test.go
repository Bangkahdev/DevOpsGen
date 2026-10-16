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

func TestValidate_DockerfilePython(t *testing.T) {
	arts := []output.Artifact{
		{
			Path: "Dockerfile",
			Content: []byte(`FROM python:3.12-alpine AS builder
WORKDIR /app
ENV VIRTUAL_ENV=/opt/venv
ENV PATH="$VIRTUAL_ENV/bin:$PATH"
RUN python -m venv $VIRTUAL_ENV
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
FROM python:3.12-alpine AS runtime
WORKDIR /app
ENV VIRTUAL_ENV=/opt/venv
ENV PATH="$VIRTUAL_ENV/bin:$PATH"
COPY --from=builder /opt/venv /opt/venv
COPY --from=builder /app /app
EXPOSE 8000
USER 1000:1000
CMD ["python", "app.py"]
`),
		},
	}
	r, err := New().Validate(arts)
	if err != nil || r != ResultValid {
		t.Fatalf("python Dockerfile should be valid: %v %v", r, err)
	}
}
