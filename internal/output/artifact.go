// Package output represents generated artifacts in memory before they are
// committed to the filesystem.
package output

// Artifact is a single file produced by the generator.
//
// Path must be relative, clean, and must not escape the target root.
type Artifact struct {
	Path    string
	Content []byte
}