//go:build embedassets

package webassets

import (
	"embed"
	"io/fs"
)

// dist embeds the frontend build. all: includes dotfiles and underscore-prefixed
// files; a missing dist directory fails the production build.
//
//go:embed all:dist
var dist embed.FS

// open removes the embedded directory prefix so assets are served from the root.
func open() (fs.FS, error) {
	return fs.Sub(dist, "dist")
}
