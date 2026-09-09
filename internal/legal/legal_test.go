package legal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// notices writes both documents into a fresh directory and returns it.
func notices(t *testing.T, privacy, imprint string) string {
	t.Helper()

	dir := t.TempDir()
	write(t, dir, privacyFile, privacy)
	write(t, dir, imprintFile, imprint)
	return dir
}

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", name, err)
	}
}

func TestNoDirectoryLeavesTheFeatureOff(t *testing.T) {
	// A default installation must not depend on any file existing.
	docs, err := Load("")
	if err != nil || docs != nil {
		t.Errorf(`Load("") = %v, %v; want nil, nil`, docs, err)
	}
}

func TestBothNoticesAreLoadedTogether(t *testing.T) {
	docs, err := Load(notices(t, "<p>privacy</p>", "<p>imprint</p>"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := string(docs.privacy); got != "<p>privacy</p>" {
		t.Errorf("privacy = %q", got)
	}
	if got := string(docs.imprint); got != "<p>imprint</p>" {
		t.Errorf("imprint = %q", got)
	}
}

func TestTheSizeLimitIsCheckedAtItsBoundary(t *testing.T) {
	for _, tc := range []struct {
		size    int
		wantErr bool
	}{
		{maxDocumentBytes, false},
		{maxDocumentBytes + 1, true},
	} {
		_, err := Load(notices(t, strings.Repeat("a", tc.size), "<p>imprint</p>"))
		if (err != nil) != tc.wantErr {
			t.Errorf("a document of %d bytes gave error %v, wantErr %v", tc.size, err, tc.wantErr)
		}
	}
}

func TestAnUnusableDocumentIsRefusedAtStartup(t *testing.T) {
	// Half a set of notices is worse than none: the app would advertise a link
	// that fails. Every one of these must stop the process instead.
	for _, tc := range []struct {
		name    string
		prepare func(t *testing.T) (dir, wantPath string)
	}{
		{"the directory does not exist", func(t *testing.T) (string, string) {
			dir := filepath.Join(t.TempDir(), "absent")
			return dir, filepath.Join(dir, privacyFile)
		}},
		{"a document is missing", func(t *testing.T) (string, string) {
			dir := t.TempDir()
			write(t, dir, privacyFile, "<p>privacy</p>")
			return dir, filepath.Join(dir, imprintFile)
		}},
		{"a document holds only whitespace", func(t *testing.T) (string, string) {
			dir := notices(t, " \n\t ", "<p>imprint</p>")
			return dir, filepath.Join(dir, privacyFile)
		}},
		{"a document is not valid UTF-8", func(t *testing.T) (string, string) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, privacyFile), []byte{0xff, 0xfe}, 0o644); err != nil {
				t.Fatalf("writing: %v", err)
			}
			write(t, dir, imprintFile, "<p>imprint</p>")
			return dir, filepath.Join(dir, privacyFile)
		}},
		{"a document is a directory", func(t *testing.T) (string, string) {
			dir := t.TempDir()
			if err := os.Mkdir(filepath.Join(dir, privacyFile), 0o755); err != nil {
				t.Fatalf("creating: %v", err)
			}
			return dir, filepath.Join(dir, privacyFile)
		}},
		{"a document is a symbolic link", func(t *testing.T) (string, string) {
			// The target is perfectly good; it is refused because following it
			// would publish a file from outside the configured directory.
			outside := filepath.Join(t.TempDir(), "elsewhere.html")
			write(t, filepath.Dir(outside), "elsewhere.html", "<p>elsewhere</p>")

			dir := t.TempDir()
			link := filepath.Join(dir, privacyFile)
			if err := os.Symlink(outside, link); err != nil {
				t.Skipf("no symbolic links on this filesystem: %v", err)
			}
			write(t, dir, imprintFile, "<p>imprint</p>")
			return dir, link
		}},
	} {
		dir, wantPath := tc.prepare(t)

		_, err := Load(dir)
		if err == nil {
			t.Errorf("%s: Load returned no error, want one", tc.name)
			continue
		}
		if !strings.Contains(err.Error(), wantPath) {
			t.Errorf("%s: error does not name %q: %v", tc.name, wantPath, err)
		}
	}
}

func TestErrorsDoNotRepeatDocumentContents(t *testing.T) {
	// A startup error goes to the container log; the operator's text must not.
	const secret = "the operator's private address"
	_, err := Load(notices(t, strings.Repeat(secret, 40000), "<p>imprint</p>"))
	if err == nil {
		t.Fatal("Load returned no error for an oversized document")
	}
	if strings.Contains(err.Error(), secret) {
		t.Errorf("the error repeats the document: %v", err)
	}
}

func TestDocumentsAreReadOnceSoEditsNeedARestart(t *testing.T) {
	dir := notices(t, "<p>first</p>", "<p>imprint</p>")
	docs, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	write(t, dir, privacyFile, "<p>second</p>")

	if got := string(docs.privacy); got != "<p>first</p>" {
		t.Errorf("the running process now serves %q; it must keep its loaded copy", got)
	}
	reloaded, err := Load(dir)
	if err != nil {
		t.Fatalf("reloading: %v", err)
	}
	if got := string(reloaded.privacy); got != "<p>second</p>" {
		t.Errorf("a restart loaded %q, want the edited text", got)
	}
}
