// Package legal loads and serves the optional privacy and legal notices an
// operator supplies as local HTML files. The feature is switched off unless a
// document directory is configured, and a default installation therefore reads
// no files and publishes no notices.
//
// The documents are read once at startup and held in memory, so no request ever
// depends on a filesystem lookup and the two notices can never be served from
// different versions of the operator's text.
package legal

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// The only two file names ever read from the configured directory. Nothing else
// in it is opened, listed or served, so the directory may hold an operator's
// working files without exposing them.
const (
	privacyFile = "privacy.html"
	imprintFile = "imprint.html"
)

// maxDocumentBytes bounds each notice at one mebibyte. The limit is not a
// judgement about how long a notice may be — prose does not approach it — but a
// guard so that a mistyped path pointing at a large file fails at startup with a
// clear message instead of being read into memory.
const maxDocumentBytes = 1 << 20

// Documents holds the notices read at startup. Its bytes are never modified
// afterwards, so every request serves the same copy without touching the disk.
// A nil *Documents means the feature is switched off.
type Documents struct {
	privacy []byte
	imprint []byte
}

// Load reads both notices from dir. An empty dir switches the feature off: it
// returns nil documents and a nil error without touching the filesystem. A
// relative dir is resolved against the process working directory.
//
// Either document being missing or unusable is an error, because half a set of
// notices is worse than none: the application would advertise a link that fails.
// The returned error names the file and the reason, never the file's contents.
func Load(dir string) (*Documents, error) {
	if dir == "" {
		return nil, nil
	}

	privacy, err := readDocument(dir, privacyFile)
	if err != nil {
		return nil, err
	}

	imprint, err := readDocument(dir, imprintFile)
	if err != nil {
		return nil, err
	}

	return &Documents{privacy: privacy, imprint: imprint}, nil
}

// readDocument reads one notice and rejects anything that is not a plain,
// nonempty UTF-8 text file within the size limit.
func readDocument(dir, name string) ([]byte, error) {
	path := filepath.Join(dir, name)

	// Lstat rather than Stat: a symbolic link is refused instead of followed, so a
	// link cannot be used to publish a file from outside the configured directory.
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read the notice %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("the notice %s is a symbolic link; it must be a regular file", path)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("the notice %s is not a regular file", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read the notice %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	// Read one byte past the limit so an oversized file is detected as such rather
	// than silently published in truncated form.
	content, err := io.ReadAll(io.LimitReader(file, maxDocumentBytes+1))
	if err != nil {
		return nil, fmt.Errorf("cannot read the notice %s: %w", path, err)
	}
	if len(content) > maxDocumentBytes {
		return nil, fmt.Errorf("the notice %s is larger than the limit of %d bytes", path, maxDocumentBytes)
	}
	if strings.TrimSpace(string(content)) == "" {
		return nil, fmt.Errorf("the notice %s is empty", path)
	}
	if !utf8.Valid(content) {
		return nil, fmt.Errorf("the notice %s is not valid UTF-8 text", path)
	}

	return content, nil
}
