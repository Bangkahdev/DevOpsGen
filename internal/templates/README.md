//go:build ignore
// +build ignore

// This file documents the template source layout. Actual templates are
// embedded via embed.FS from the parent templates/ directory at build time.
//
// Layout:
//   templates/docker/        Dockerfile templates
//   templates/github-actions/ GitHub Actions workflow templates
//   templates/kubernetes/    Kubernetes manifest templates