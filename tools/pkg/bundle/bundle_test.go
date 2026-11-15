package bundle_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"hash/crc32"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/galigaribaldi/Memorandum-tex/tools/internal/fixturetest"
	"github.com/galigaribaldi/Memorandum-tex/tools/pkg/bundle"
	"github.com/galigaribaldi/Memorandum-tex/tools/pkg/contract"
)

func buildOptions(t *testing.T, templateDir string) bundle.BuildOptions {
	return bundle.BuildOptions{
		TemplateDir: templateDir,
		SharedDir:   t.TempDir(),
		LicensesDir: filepath.Join(fixturetest.RepoRoot(t), "LICENSES"),
	}
}

func okMinimaDir(t *testing.T) string {
	return filepath.Join(fixturetest.DataDir(t), "templates", "prueba", "ok-minima")
}

type caseEntry struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Zeros   int    `json:"zeros"`
	Symlink string `json:"symlink"`
}

type bundleCase struct {
	Name        string      `json:"name"`
	From        string      `json:"from"`
	Entries     []caseEntry `json:"entries"`
	Omit        []string    `json:"omit"`
	Rename      string      `json:"rename"`
	IndexSHA256 string      `json:"indexSha256"`
	Expect      string      `json:"expect"`
}

var expectedErrors = map[string]error{
	"ErrChecksum": contract.ErrChecksum, "ErrUnsafePath": contract.ErrUnsafePath,
	"ErrForbiddenFile": contract.ErrForbiddenFile, "ErrBundleLayout": contract.ErrBundleLayout,
	"ErrTooLarge": contract.ErrTooLarge,
}

// caseZip builds the zip a case describes.
func caseZip(t *testing.T, c bundleCase) []byte {
	t.Helper()
	var entries []bundle.Entry
	if c.From != "" {
		dir := filepath.Join(fixturetest.DataDir(t), "bundles", filepath.FromSlash(c.From))
		var err error
		if entries, _, err = bundle.Collect(buildOptions(t, dir)); err != nil {
			t.Fatal(err)
		}
	}
	omit := map[string]bool{}
	for _, o := range c.Omit {
		omit[o] = true
	}
	kept := entries[:0]
	for _, e := range entries {
		if omit[e.Name] {
			delete(omit, e.Name)
			continue
		}
		if c.Rename != "" {
			_, rest, _ := strings.Cut(e.Name, "/")
			e.Name = c.Rename + "/" + rest
		}
		kept = append(kept, e)
	}
	if len(omit) > 0 {
		t.Fatalf("omit nombra entradas que no existen: %v", omit)
	}
	entries = kept
	for _, ce := range c.Entries {
		e := bundle.Entry{Name: ce.Path, Data: []byte(ce.Content), Mode: 0o644}
		switch {
		case ce.Zeros > 0:
			e.Data = make([]byte, ce.Zeros)
		case ce.Symlink != "":
			e.Data, e.Mode = []byte(ce.Symlink), fs.ModeSymlink|0o777
		}
		entries = append(entries, e)
	}
	var buf bytes.Buffer
	if err := bundle.WriteZip(&buf, entries); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// TestBundleCases runs every case of bundles/cases.json.
func TestBundleCases(t *testing.T) {
	var file struct {
		Cases []bundleCase `json:"cases"`
	}
	fixturetest.ReadJSON(t, "bundles/cases.json", &file)
	if len(file.Cases) != 16 {
		t.Fatalf("bundles/cases.json tiene %d casos; la documentación dice 16", len(file.Cases))
	}
	for _, c := range file.Cases {
		t.Run(c.Name, func(t *testing.T) {
			data := caseZip(t, c)
			sum, _, _ := bundle.SHA256(bytes.NewReader(data))
			want := sum
			if c.IndexSHA256 != "" {
				want = c.IndexSHA256
			}
			err := bundle.VerifySHA256(bytes.NewReader(data), want)
			if err == nil {
				dest := filepath.Join(t.TempDir(), "destino")
				var res *bundle.Result
				res, err = bundle.SafeExtract(context.Background(), bytes.NewReader(data), int64(len(data)), dest, bundle.ExtractOptions{})
				if err == nil {
					checkExtracted(t, res)
				} else if _, statErr := os.Stat(dest); statErr == nil {
					t.Errorf("tras un error no debe quedar nada en el destino")
				}
			}
			if c.Expect == "ok" {
				if err != nil {
					t.Fatalf("se esperaba éxito: %v", err)
				}
				return
			}
			wantErr, ok := expectedErrors[c.Expect]
			if !ok {
				t.Fatalf("error esperado desconocido %q", c.Expect)
			}
			if !errors.Is(err, wantErr) {
				t.Fatalf("se esperaba %s, dio %v", c.Expect, err)
			}
		})
	}
}

func checkExtracted(t *testing.T, res *bundle.Result) {
	t.Helper()
	for _, f := range []string{"template.json", "LICENSE", "README.md", ".creative-ink-meta", res.MainFile} {
		if _, err := os.Stat(filepath.Join(res.Dir, filepath.FromSlash(f))); err != nil {
			t.Errorf("falta %s: %v", f, err)
		}
	}
	lic, _ := os.ReadFile(filepath.Join(res.Dir, "LICENSE"))
	if !strings.HasPrefix(string(lic), "MIT No Attribution") {
		t.Errorf("LICENSE debe ser el texto de MIT-0: %q", firstLine(lic))
	}
	meta, _ := os.ReadFile(filepath.Join(res.Dir, ".creative-ink-meta"))
	var m struct {
		Version  int    `json:"version"`
		MainFile string `json:"mainFile"`
	}
	if json.Unmarshal(meta, &m) != nil || m.Version != 1 || m.MainFile != "main.tex" {
		t.Errorf(".creative-ink-meta = %s", meta)
	}
	tj, _ := os.ReadFile(filepath.Join(res.Dir, "template.json"))
	if bytes.Contains(tj, []byte(`"$schema"`)) {
		t.Error("el template.json del bundle no debe llevar $schema")
	}
}

func firstLine(b []byte) string { s, _, _ := strings.Cut(string(b), "\n"); return s }

// TestReproducible: the same template gives byte-identical zips (B8).
func TestReproducible(t *testing.T) {
	var a, b bytes.Buffer
	if _, err := bundle.Build(&a, buildOptions(t, okMinimaDir(t))); err != nil {
		t.Fatal(err)
	}
	if _, err := bundle.Build(&b, buildOptions(t, okMinimaDir(t))); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a.Bytes(), b.Bytes()) {
		t.Fatal("dos compilaciones del mismo bundle difieren")
	}
	zr, err := zip.NewReader(bytes.NewReader(a.Bytes()), int64(a.Len()))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, f := range zr.File {
		if !f.Modified.Equal(bundle.FixedTime) {
			t.Errorf("%s tiene fecha %v", f.Name, f.Modified)
		}
		names = append(names, f.Name)
	}
	for i := 1; i < len(names); i++ {
		if names[i-1] >= names[i] {
			t.Errorf("entradas desordenadas: %s antes de %s", names[i-1], names[i])
		}
	}
}

func TestComponentsGoUnderLatexInk(t *testing.T) {
	opt := buildOptions(t, okMinimaDir(t))
	if err := os.MkdirAll(filepath.Join(opt.SharedDir, "motor"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(opt.SharedDir, "motor", "ink-motor.tex"), []byte("% motor\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "prueba", "ok-minima")
	fixturetest.CopyDir(t, okMinimaDir(t), dir)
	tj := filepath.Join(dir, "template.json")
	data, _ := os.ReadFile(tj)
	if err := os.WriteFile(tj, []byte(strings.Replace(string(data), `"modes":`, `"components": ["motor"], "modes":`, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	opt.TemplateDir = dir
	entries, _, err := bundle.Collect(opt)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range entries {
		found = found || e.Name == "ok-minima/Latex/ink/motor/ink-motor.tex"
	}
	if !found {
		t.Error("el componente motor debe quedar en Latex/ink/motor/")
	}
}

func TestDestMustBeEmpty(t *testing.T) {
	var buf bytes.Buffer
	if _, err := bundle.Build(&buf, buildOptions(t, okMinimaDir(t))); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	if err := os.WriteFile(filepath.Join(dest, "mio.tex"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := bundle.SafeExtract(context.Background(), bytes.NewReader(buf.Bytes()), int64(buf.Len()), dest, bundle.ExtractOptions{})
	if !errors.Is(err, contract.ErrDestExists) {
		t.Fatalf("un destino con archivos nunca se sobrescribe: %v", err)
	}
	if data, _ := os.ReadFile(filepath.Join(dest, "mio.tex")); string(data) != "x" {
		t.Error("el archivo del usuario cambió")
	}
	empty := t.TempDir() // exists and is empty: allowed
	if _, err := bundle.SafeExtract(context.Background(), bytes.NewReader(buf.Bytes()), int64(buf.Len()), empty, bundle.ExtractOptions{}); err != nil {
		t.Fatalf("un destino vacío debe aceptarse: %v", err)
	}
}

func TestExpectIDAndVersion(t *testing.T) {
	var buf bytes.Buffer
	if _, err := bundle.Build(&buf, buildOptions(t, okMinimaDir(t))); err != nil {
		t.Fatal(err)
	}
	for _, opt := range []bundle.ExtractOptions{{ExpectID: "otra"}, {ExpectVersion: "9.9.9"}} {
		dest := filepath.Join(t.TempDir(), "d")
		if _, err := bundle.SafeExtract(context.Background(), bytes.NewReader(buf.Bytes()), int64(buf.Len()), dest, opt); !errors.Is(err, contract.ErrBundleLayout) {
			t.Errorf("%+v: el bundle debe ser el que anunció el índice: %v", opt, err)
		}
	}
}

func TestCanceled(t *testing.T) {
	var buf bytes.Buffer
	if _, err := bundle.Build(&buf, buildOptions(t, okMinimaDir(t))); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	dest := filepath.Join(t.TempDir(), "d")
	if _, err := bundle.SafeExtract(ctx, bytes.NewReader(buf.Bytes()), int64(buf.Len()), dest, bundle.ExtractOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("se esperaba context.Canceled: %v", err)
	}
	if _, err := os.Stat(dest); err == nil {
		t.Error("una extracción cancelada no debe dejar el destino")
	}
	leftovers, _ := filepath.Glob(filepath.Join(filepath.Dir(dest), ".ink-extract-*"))
	if len(leftovers) > 0 {
		t.Errorf("quedó la carpeta temporal: %v", leftovers)
	}
}

// TestLyingSize: a zip whose header declares a small size but whose data
// inflates to much more must fail while extracting, not fill the disk.
func TestLyingSize(t *testing.T) {
	entries, _, err := bundle.Collect(buildOptions(t, okMinimaDir(t)))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		w, _ := zw.CreateHeader(&zip.FileHeader{Name: e.Name, Method: zip.Deflate, Modified: bundle.FixedTime})
		_, _ = w.Write(e.Data)
	}
	// Raw entry: real content 2 MB of zeros, declared as 10 bytes.
	var deflated bytes.Buffer
	fw := zip.NewWriter(&deflated)
	big := make([]byte, 2<<20)
	tmp, _ := fw.Create("x")
	_, _ = tmp.Write(big)
	_ = fw.Close()
	inner, _ := zip.NewReader(bytes.NewReader(deflated.Bytes()), int64(deflated.Len()))
	rawData, _ := inner.File[0].OpenRaw()
	raw := new(bytes.Buffer)
	_, _ = raw.ReadFrom(rawData)
	h := &zip.FileHeader{Name: "ok-minima/mentira.txt", Method: zip.Deflate, Modified: bundle.FixedTime,
		CompressedSize64: uint64(raw.Len()), UncompressedSize64: 10, CRC32: crc32.ChecksumIEEE(big[:10])}
	rw, err := zw.CreateRaw(h)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = rw.Write(raw.Bytes())
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "d")
	_, err = bundle.SafeExtract(context.Background(), bytes.NewReader(buf.Bytes()), int64(buf.Len()), dest, bundle.ExtractOptions{})
	if err == nil || !(errors.Is(err, contract.ErrTooLarge) || errors.Is(err, contract.ErrBundleLayout)) {
		t.Fatalf("un tamaño falso debe fallar: %v", err)
	}
	if info, statErr := os.Stat(filepath.Join(dest, "mentira.txt")); statErr == nil && info.Size() > 11 {
		t.Errorf("se escribieron %d bytes de un archivo que declaraba 10", info.Size())
	}
}

func TestShellEscapeBundleRefused(t *testing.T) {
	entries, _, err := bundle.Collect(buildOptions(t, okMinimaDir(t)))
	if err != nil {
		t.Fatal(err)
	}
	for i, e := range entries {
		if path.Base(e.Name) == "template.json" {
			entries[i].Data = bytes.Replace(e.Data, []byte(`"defaultEngine"`), []byte(`"shellEscape": true, "defaultEngine"`), 1)
		}
	}
	var buf bytes.Buffer
	if err := bundle.WriteZip(&buf, entries); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "d")
	if _, err := bundle.SafeExtract(context.Background(), bytes.NewReader(buf.Bytes()), int64(buf.Len()), dest, bundle.ExtractOptions{}); !errors.Is(err, contract.ErrForbiddenFile) {
		t.Fatalf("un bundle con shellEscape debe rechazarse: %v", err)
	}
}
