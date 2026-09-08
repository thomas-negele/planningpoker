package game

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// parsePackage parses this package's own source, excluding the tests. The tests
// themselves are allowed to import whatever they need to do the checking.
func parsePackage(t *testing.T) (*token.FileSet, []*ast.File) {
	t.Helper()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading the package directory: %v", err)
	}

	fset := token.NewFileSet()
	var files []*ast.File
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(".", name), nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parsing %s: %v", name, err)
		}
		files = append(files, f)
	}

	if len(files) == 0 {
		t.Fatal("no source files found; the check below would pass vacuously")
	}
	return fset, files
}

// TestPackageImportsNoTransport enforces the rule that gives this package its
// value: the rules of the game must be readable and testable without a network.
//
// The constraint is easy to state and easy to violate — one import added for a
// single convenient helper is all it takes — so it is checked rather than trusted.
func TestPackageImportsNoTransport(t *testing.T) {
	forbidden := map[string]string{
		"net/http":                   "the rules must not know how they are served",
		"net":                        "the rules must not reach the network at all",
		"github.com/coder/websocket": "the rules must not know how they are transmitted",
		"encoding/json":              "serialisation belongs to the transport layer, not to the rules",
		"time":                       "room lifetime is a matter for the layer that owns a clock",
		"sync":                       "each room is owned by a single goroutine; a lock here would mean that model was abandoned",
	}

	_, files := parsePackage(t)
	for _, f := range files {
		for _, spec := range f.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("unquoting import %s: %v", spec.Path.Value, err)
			}
			if reason, bad := forbidden[path]; bad {
				t.Errorf("%s imports %q, which this package must not: %s", f.Name.Name, path, reason)
			}
		}
	}
}

// TestPackageHasNoConcurrency enforces the other half of the same idea. Every room
// is owned by exactly one goroutine in the layer above, and nothing outside that
// goroutine touches its state. A channel, a go statement or a mutex appearing here
// would be a sign that the single-owner model had quietly been abandoned somewhere
// else, and the right response would be to fix that rather than to lock this.
func TestPackageHasNoConcurrency(t *testing.T) {
	fset, files := parsePackage(t)

	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.GoStmt:
				t.Errorf("%s: a go statement — this package must stay single-threaded",
					fset.Position(node.Pos()))
			case *ast.ChanType:
				t.Errorf("%s: a channel type — this package must stay single-threaded",
					fset.Position(node.Pos()))
			case *ast.SelectStmt:
				t.Errorf("%s: a select statement — this package must stay single-threaded",
					fset.Position(node.Pos()))
			}
			return true
		})
	}
}
