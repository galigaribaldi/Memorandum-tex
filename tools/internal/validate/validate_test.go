package validate_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/galigaribaldi/Memorandum-tex/tools/internal/fixturetest"
	"github.com/galigaribaldi/Memorandum-tex/tools/internal/validate"
	"github.com/galigaribaldi/Memorandum-tex/tools/pkg/contract"
)

func options(t *testing.T) validate.Options {
	return validate.Options{
		SharedDir:   t.TempDir(),
		LicensesDir: filepath.Join(fixturetest.RepoRoot(t), "LICENSES"),
	}
}

type findingKey struct {
	code  string
	level contract.Level
}

func keysOf(r contract.Report) map[findingKey]bool {
	out := map[findingKey]bool{}
	for _, f := range r.Findings {
		out[findingKey{f.Code, f.Level}] = true
	}
	return out
}

// copyOkMinima copies the base template to <tmp>/prueba/ok-minima, so the
// folder and its parent keep the names the manifest expects (V04).
func copyOkMinima(t *testing.T) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "prueba", "ok-minima")
	fixturetest.CopyDir(t, filepath.Join(fixturetest.DataDir(t), "templates", "prueba", "ok-minima"), dst)
	return dst
}

func TestOkMinimaFolderIsValid(t *testing.T) {
	for name, dir := range map[string]string{
		"en el repositorio": filepath.Join(fixturetest.DataDir(t), "templates", "prueba", "ok-minima"),
		"copia":             copyOkMinima(t),
	} {
		rep := validate.Folder(dir, options(t))
		if rep.HasErrors() {
			t.Errorf("%s: ok-minima debe ser válida: %v", name, rep.Findings)
		}
		for _, f := range rep.Warnings() {
			if f.Code != "V06" {
				t.Errorf("%s: aviso inesperado %v", name, f)
			}
		}
	}
}

type op struct {
	Op      string `json:"op"`
	Path    string `json:"path"`
	Content string `json:"content"`
	Old     string `json:"old"`
	New     string `json:"new"`
	Target  string `json:"target"`
	Dir     string `json:"dir"`
	BOM     bool   `json:"bom"`
	CRLF    bool   `json:"crlf"`
	Latin1  bool   `json:"latin1"`
	Bytes   int64  `json:"bytes"`
	Count   int    `json:"count"`
}

type folderCase struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Ops  []op   `json:"ops"`
}

// TestInvalidFolders runs every case of invalid/folders.json: each must
// produce its finding and no other beyond those ok-minima already has.
func TestInvalidFolders(t *testing.T) {
	var file struct {
		Cases []folderCase `json:"cases"`
	}
	fixturetest.ReadJSON(t, "invalid/folders.json", &file)
	base := keysOf(validate.Folder(copyOkMinima(t), options(t)))

	covered, names := map[string]bool{}, map[string]bool{}
	for _, c := range file.Cases {
		if names[c.Name] {
			t.Errorf("nombre de caso repetido: %s", c.Name)
		}
		names[c.Name] = true
		covered[c.Code] = true
		t.Run(c.Code+"/"+c.Name, func(t *testing.T) {
			dir := copyOkMinima(t)
			for _, o := range c.Ops {
				apply(t, dir, o)
			}
			if c.Name == "colision-mayusculas" && caseInsensitive(t, dir) {
				t.Skip("este sistema de archivos no distingue mayúsculas: el segundo archivo sobrescribe al primero (B10 se prueba en pkg/bundle en todos los sistemas)")
			}
			rep := validate.Folder(dir, options(t))
			want := findingKey{c.Code, contract.LevelError}
			got := keysOf(rep)
			for k := range base {
				delete(got, k)
			}
			if !got[want] {
				t.Errorf("falta %s; hallazgos: %v", c.Code, rep.Findings)
			}
			for k := range got {
				if k != want {
					t.Errorf("hallazgo inesperado %s %s; hallazgos: %v", k.code, k.level, rep.Findings)
				}
			}
		})
	}
	for n := 30; n <= 39; n++ {
		if code := fmt.Sprintf("V%02d", n); !covered[code] {
			t.Errorf("la regla %s no tiene ningún caso en invalid/folders.json", code)
		}
	}
}

func apply(t *testing.T, dir string, o op) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(o.Path))
	switch o.Op {
	case "delete":
		if err := os.Remove(p); err != nil {
			t.Fatal(err)
		}
	case "write":
		content := o.Content
		if o.CRLF {
			content = strings.ReplaceAll(content, "\n", "\r\n")
		}
		data := []byte(content)
		if o.Latin1 {
			data = data[:0]
			for _, r := range content {
				if r > 0xFF {
					t.Fatalf("%q no cabe en Latin-1", r)
				}
				data = append(data, byte(r))
			}
		}
		if o.BOM {
			data = append([]byte("\xef\xbb\xbf"), data...)
		}
		mustWrite(t, p, data)
	case "replace":
		old, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		// A replace that finds nothing would leave the case testing nothing.
		if !strings.Contains(string(old), o.Old) {
			t.Fatalf("el fixture ya no contiene %q en %s: actualiza invalid/folders.json", o.Old, o.Path)
		}
		mustWrite(t, p, []byte(strings.Replace(string(old), o.Old, o.New, 1)))
	case "symlink":
		if err := os.Symlink(o.Target, p); err != nil {
			t.Skipf("no se pudo crear el enlace simbólico (¿Windows sin permiso?): %v", err)
		}
	case "fill":
		// Sparse: the size is real for os.Stat but uses no disk.
		f, err := os.Create(p)
		if err == nil {
			err = f.Truncate(o.Bytes)
			f.Close()
		}
		if err != nil {
			if os.MkdirAll(filepath.Dir(p), 0o755) == nil {
				if f, err = os.Create(p); err == nil {
					err = f.Truncate(o.Bytes)
					f.Close()
				}
			}
			if err != nil {
				t.Fatal(err)
			}
		}
	case "copies":
		d := filepath.Join(dir, filepath.FromSlash(o.Dir))
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < o.Count; i++ {
			mustWrite(t, filepath.Join(d, fmt.Sprintf("%04d.tex", i)), nil)
		}
	default:
		t.Fatalf("operación desconocida %q", o.Op)
	}
}

func mustWrite(t *testing.T, p string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// caseInsensitive reports whether the file system under dir treats "A" and "a" as one name.
func caseInsensitive(t *testing.T, dir string) bool {
	probe := filepath.Join(dir, "Mayusculas.probe")
	mustWrite(t, probe, nil)
	defer os.Remove(probe)
	_, err := os.Stat(filepath.Join(dir, "mayusculas.probe"))
	return err == nil
}

func TestStripCommentKeepsEscapedPercent(t *testing.T) {
	dir := copyOkMinima(t)
	// "\%" is an escaped percent, not a comment: the \input after it must still be seen.
	mustWrite(t, filepath.Join(dir, "src", "secciones", "introduccion.tex"), []byte("100\\% \\input{../fuera}\n"))
	if rep := validate.Folder(dir, options(t)); !keysOf(rep)[findingKey{"V35", contract.LevelError}] {
		t.Errorf(`\%% no debe ocultar lo que sigue: %v`, rep.Findings)
	}
	mustWrite(t, filepath.Join(dir, "src", "secciones", "introduccion.tex"), []byte("% \\input{../fuera}\n"))
	if rep := validate.Folder(dir, options(t)); keysOf(rep)[findingKey{"V35", contract.LevelError}] {
		t.Errorf("una línea comentada no debe dar V35: %v", rep.Findings)
	}
}

func TestComponentsCountTowardsLimits(t *testing.T) {
	dir := copyOkMinima(t)
	shared := t.TempDir()
	comp := filepath.Join(shared, "paletas")
	if err := os.MkdirAll(comp, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < contract.MaxBundleFiles; i++ {
		mustWrite(t, filepath.Join(comp, fmt.Sprintf("%04d.tex", i)), nil)
	}
	manifest := filepath.Join(dir, "template.json")
	data, _ := os.ReadFile(manifest)
	mustWrite(t, manifest, []byte(strings.Replace(string(data), `"modes":`, `"components": ["paletas"], "modes":`, 1)))
	rep := validate.Folder(dir, validate.Options{SharedDir: shared, LicensesDir: filepath.Join(fixturetest.RepoRoot(t), "LICENSES")})
	if !keysOf(rep)[findingKey{"V34", contract.LevelError}] {
		t.Errorf("los archivos de los componentes cuentan para el límite: %v", rep.Findings)
	}
}
