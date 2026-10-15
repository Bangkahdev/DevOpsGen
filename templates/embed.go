// Package templates embeds the source templates at build time.
//
// The embed directive must live in the same directory tree as the embedded
// files, so it lives here at the repository root alongside the template
// sources. The loader API is exported from internal/templates.
package templates

import (
	"embed"
	"io/fs"
)

//go:embed docker github-actions kubernetes
var Embed embed.FS

// Root returns the embedded template filesystem rooted at the templates
// directory.
func Root() fs.FS {
	sub, err := fs.Sub(Embed, ".")
	if err != nil {
		panic("templates: embed sub: " + err.Error())
	}
	return sub
}
