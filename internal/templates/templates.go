// Package templates exposes the bundled template loader.
//
// Templates are embedded at build time and remain available without any
// external template directory. Template version tracks the binary version
// on the MVP.
package templates

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"devopsgen.dev/devopsgen/templates"
)

// Version reports the bundled template version.
var Version = "dev"

// FS returns the embedded template filesystem rooted at the templates
// directory. Callers must treat returned paths as relative to that root.
func FS() fs.FS {
	return templates.Root()
}

// cleanPath normalizes a template path to forward slashes and rejects
// traversal. It deliberately avoids filepath.Clean, which on Windows
// converts "/" to "\" and breaks the embedded FS lookups.
func cleanPath(name string) (string, error) {
	name = filepath.ToSlash(filepath.Clean(name))
	name = strings.ReplaceAll(name, "\\", "/")
	if name == "." || name == ".." || strings.HasPrefix(name, "../") || strings.HasPrefix(name, "/") {
		return "", fmt.Errorf("templates: path %q escapes root", name)
	}
	return name, nil
}

// Read returns the embedded template content for the given relative path.
// Path is normalized and must stay within the embedded root.
func Read(name string) ([]byte, error) {
	n, err := cleanPath(name)
	if err != nil {
		return nil, err
	}
	return fs.ReadFile(FS(), n)
}

// List returns all template paths under the given subdirectory.
// Returned paths use forward slashes for cross-platform consistency.
func List(dir string) ([]string, error) {
	d, err := cleanPath(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	err = fs.WalkDir(FS(), d, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		out = append(out, filepath.ToSlash(path))
		return nil
	})
	return out, err
}

// Exists reports whether the named template is present.
func Exists(name string) bool {
	b, err := Read(name)
	return err == nil && len(b) > 0
}
