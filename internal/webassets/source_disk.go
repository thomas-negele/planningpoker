//go:build !embedassets

package webassets

import (
	"fmt"
	"io/fs"
	"os"
)

// diskDir is resolved relative to the process working directory.
const diskDir = "internal/webassets/dist"

// open reads a previous frontend build from disk. Development through Vite
// does not require this directory; direct asset requests return 503 if it is absent.
func open() (fs.FS, error) {
	info, err := os.Stat(diskDir)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("%w at %s", errNoBuild, diskDir)
	}
	return os.DirFS(diskDir), nil
}
