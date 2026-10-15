package generator

import (
	"strings"
	"testing"

	"devopsgen.dev/devopsgen/internal/config"
)

func TestGenerate_PythonFastAPI(t *testing.T) {
	cfg := config.Config{
		Stack:    config.StackPython,
		AppType:  config.AppTypeFastAPI,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryDockerHub,
		Image:    "bangkah/api",
		Port:     8000,
		StartCmd: "uvicorn app:app",
	}
	arts, err := New().Generate(cfg)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	var dockerfile []byte
	for _, a := range arts {
		if a.Path == "Dockerfile" {
			dockerfile = a.Content
		}
	}
	if dockerfile == nil {
		t.Fatal("Dockerfile not generated for python")
	}
	if !strings.Contains(string(dockerfile), "python:3.12-slim") {
		t.Fatal("Dockerfile missing python base image")
	}
	if !strings.Contains(string(dockerfile), "EXPOSE 8000") {
		t.Fatal("Dockerfile missing EXPOSE 8000")
	}
	if !strings.Contains(string(dockerfile), "uvicorn app:app") {
		t.Fatal("Dockerfile missing uvicorn start command")
	}
}

func TestGenerate_PythonDjango(t *testing.T) {
	cfg := config.Config{
		Stack:    config.StackPython,
		AppType:  config.AppTypeDjango,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryGHCR,
		Image:    "bangkah/api",
		Port:     8000,
		BuildCmd: "python manage.py collectstatic --noinput",
		StartCmd: "python manage.py runserver",
	}
	arts, err := New().Generate(cfg)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	var dockerfile []byte
	for _, a := range arts {
		if a.Path == "Dockerfile" {
			dockerfile = a.Content
		}
	}
	if dockerfile == nil {
		t.Fatal("Dockerfile not generated for django")
	}
	if !strings.Contains(string(dockerfile), "collectstatic") {
		t.Fatal("Dockerfile missing collectstatic build command")
	}
	if !strings.Contains(string(dockerfile), "python manage.py runserver") {
		t.Fatal("Dockerfile missing django start command")
	}
}

func TestGenerate_PythonFlask(t *testing.T) {
	cfg := config.Config{
		Stack:    config.StackPython,
		AppType:  config.AppTypeFlask,
		CI:       config.CIGitHubActions,
		Registry: config.RegistryGHCR,
		Image:    "bangkah/api",
		Port:     5000,
	}
	arts, err := New().Generate(cfg)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	var dockerfile []byte
	for _, a := range arts {
		if a.Path == "Dockerfile" {
			dockerfile = a.Content
		}
	}
	if dockerfile == nil {
		t.Fatal("Dockerfile not generated for flask")
	}
	if !strings.Contains(string(dockerfile), "EXPOSE 5000") {
		t.Fatal("Dockerfile missing EXPOSE 5000")
	}
}
