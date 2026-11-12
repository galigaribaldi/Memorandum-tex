// Package fixturetest holds helpers shared by the tests of this module: where
// the fixtures and the repository are on disk, how to read a case file and how
// to apply a JSON Merge Patch (RFC 7386) to a manifest.
package fixturetest

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/galigaribaldi/Memorandum-tex/tools/fixtures"
)

// DataDir is the absolute path of tools/fixtures/data.
func DataDir(t testing.TB) string {
	t.Helper()
	dir, ok := fixtures.Dir()
	if !ok {
		t.Fatal("fixtures are not on disk")
	}
	return dir
}

// RepoRoot is the root of the Memorandum-tex repository (it holds LICENSES/).
func RepoRoot(t testing.TB) string {
	t.Helper()
	root := filepath.Clean(filepath.Join(DataDir(t), "..", "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "LICENSES")); err != nil {
		t.Fatalf("repository root not found at %s: %v", root, err)
	}
	return root
}

// ReadJSON decodes the fixture at path (relative to data/, with "/") into v,
// keeping numbers as json.Number so a round trip does not change them.
func ReadJSON(t testing.TB, path string, v any) {
	t.Helper()
	raw, err := fs.ReadFile(fixtures.FS(), path)
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(v); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}

// MergePatch applies an RFC 7386 JSON Merge Patch: objects are merged
// recursively, a null value deletes the key and anything else (lists
// included) replaces the target. target is modified in place when it is an
// object.
func MergePatch(target, patch any) any {
	p, ok := patch.(map[string]any)
	if !ok {
		return patch
	}
	t, ok := target.(map[string]any)
	if !ok {
		t = map[string]any{}
	}
	for k, v := range p {
		if v == nil {
			delete(t, k)
			continue
		}
		t[k] = MergePatch(t[k], v)
	}
	return t
}
