// Package templates loads the bundled templates via embed.FS.
//
// Templates are embedded at build time and remain available without any
// external template directory. Template version tracks the binary version
// on the MVP.
package templates

// Version reports the bundled template version.
const Version = "dev"