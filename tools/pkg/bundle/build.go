// Package bundle builds, verifies and safely extracts template bundles
// (docs/contrato-plantillas.md §9, rules B1–B10).
//
// Build and WriteZip are used by inkt and by catalogtest. VerifySHA256 and
// SafeExtract are what Creative-Ink calls on a downloaded zip: SafeExtract
// checks every entry before writing anything and never trusts the sizes the
// zip declares.
package bundle

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/galigaribaldi/Memorandum-tex/tools/pkg/contract"
)

// FixedTime is the modification time of every entry (B8): with it, the same
// content gives the same zip, and the same sha256, on any OS (with the same Go
// version, whose Deflate output is stable).
var FixedTime = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

// ComponentsDir is where inkt bundle copies the declared components:
// Latex/ink/<component>/… (for example Latex/ink/motor/ink-motor.tex).
const ComponentsDir = "Latex/ink"

// Entry is one file of a bundle. Name is the full path inside the zip, with
// "/" and the root folder ("ok-minima/main.tex"). A Mode with fs.ModeSymlink
// makes Data the link target (only tests build such entries).
type Entry struct {
	Name string
	Data []byte
	Mode fs.FileMode
}

// BuildOptions locate a template of the repository and what it needs.
type BuildOptions struct {
	TemplateDir string // templates/<categoria>/<id>/
	SharedDir   string // shared/, for components
	LicensesDir string // LICENSES/, for the LICENSE file
}

// Collect returns the entries of a template's bundle, sorted by name, and its
// manifest. The template must be valid (run inkt validate first); Collect
// still refuses what would break the contract.
func Collect(opt BuildOptions) ([]Entry, *contract.Manifest, error) {
	data, err := os.ReadFile(filepath.Join(opt.TemplateDir, "template.json"))
	if err != nil {
		return nil, nil, err
	}
	m, rep := contract.ValidateManifest(data, contract.Options{Origin: contract.OriginRepository})
	if m == nil || rep.HasErrors() {
		return nil, nil, fmt.Errorf("%w: template.json inválido; ejecuta inkt validate: %v", contract.ErrBundleLayout, rep.Errors())
	}
	root := m.ID + "/"
	var entries []Entry
	add := func(name string, data []byte) {
		entries = append(entries, Entry{Name: root + name, Data: data, Mode: 0o644})
	}

	// template.json without $schema (B6), in a stable format.
	pub := *m
	pub.Schema = ""
	tj, err := json.MarshalIndent(pub, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	add("template.json", append(tj, '\n'))

	lic, err := licenseText(m.License, opt.LicensesDir)
	if err != nil {
		return nil, nil, err
	}
	add(contract.LicenseFileName, lic)

	readme, err := os.ReadFile(filepath.Join(opt.TemplateDir, "README.md"))
	if err != nil {
		return nil, nil, fmt.Errorf("%w: falta README.md", contract.ErrBundleLayout)
	}
	add("README.md", readme)

	meta, _ := json.MarshalIndent(struct {
		Version  int    `json:"version"`
		MainFile string `json:"mainFile"`
	}{1, m.Entry}, "", "  ")
	add(contract.MetaFileName, append(meta, '\n'))

	src := filepath.Join(opt.TemplateDir, "src")
	err = walkFiles(src, func(rel string, data []byte) error {
		if strings.HasPrefix(rel, ComponentsDir+"/") {
			return fmt.Errorf("%w: src/%s: %s/ está reservado para los componentes", contract.ErrBundleLayout, rel, ComponentsDir)
		}
		add(rel, data)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	for _, c := range m.Components {
		dir := filepath.Join(opt.SharedDir, filepath.FromSlash(c))
		err := walkFiles(dir, func(rel string, data []byte) error {
			add(ComponentsDir+"/"+c+"/"+rel, data)
			return nil
		})
		if err != nil {
			return nil, nil, fmt.Errorf("componente %s: %w", c, err)
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries, m, nil
}

// Build writes the bundle of a template to w and returns its manifest.
func Build(w io.Writer, opt BuildOptions) (*contract.Manifest, error) {
	entries, m, err := Collect(opt)
	if err != nil {
		return nil, err
	}
	return m, WriteZip(w, entries)
}

// WriteZip writes entries as a reproducible zip (B8): in the given order,
// with FixedTime and Deflate. Collect already sorts them.
func WriteZip(w io.Writer, entries []Entry) error {
	zw := zip.NewWriter(w)
	for _, e := range entries {
		h := &zip.FileHeader{Name: e.Name, Method: zip.Deflate, Modified: FixedTime}
		mode := e.Mode
		if mode == 0 {
			mode = 0o644
		}
		if mode&fs.ModeSymlink != 0 {
			h.Method = zip.Store
		}
		h.SetMode(mode)
		f, err := zw.CreateHeader(h)
		if err != nil {
			return err
		}
		if _, err := f.Write(e.Data); err != nil {
			return err
		}
	}
	return zw.Close()
}

// licenseText joins LICENSES/<id>.txt for every identifier of the
// expression, in order, separated by a blank line (B6).
func licenseText(expr, dir string) ([]byte, error) {
	ids, ok := contract.ParseLicense(expr)
	if !ok {
		return nil, fmt.Errorf("%w: licencia %q inválida", contract.ErrBundleLayout, expr)
	}
	var b bytes.Buffer
	for i, id := range ids {
		text, err := os.ReadFile(filepath.Join(dir, id+".txt"))
		if err != nil {
			return nil, fmt.Errorf("%w: falta LICENSES/%s.txt", contract.ErrBundleLayout, id)
		}
		if i > 0 {
			b.WriteString("\n")
		}
		b.Write(bytes.TrimRight(text, "\n"))
		b.WriteString("\n")
	}
	return b.Bytes(), nil
}

// walkFiles calls fn for every regular file under root with its path
// relative to root ("/"), refusing what the contract forbids.
func walkFiles(root string, fn func(rel string, data []byte) error) error {
	return filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		r, _ := filepath.Rel(root, p)
		rel := filepath.ToSlash(r)
		switch {
		case d.Type()&fs.ModeSymlink != 0 || !d.Type().IsRegular():
			return fmt.Errorf("%w: %s no es un archivo normal", contract.ErrUnsafePath, rel)
		case !contract.ValidRelPath(rel):
			return fmt.Errorf("%w: %s", contract.ErrUnsafePath, rel)
		case !contract.AllowedSourceFile(d.Name()):
			return fmt.Errorf("%w: %s", contract.ErrForbiddenFile, rel)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return fn(rel, data)
	})
}
