package contract_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/galigaribaldi/Memorandum-tex/tools/fixtures"
	"github.com/galigaribaldi/Memorandum-tex/tools/internal/fixturetest"
	"github.com/galigaribaldi/Memorandum-tex/tools/pkg/contract"
)

const okMinima = "templates/prueba/ok-minima/template.json"

func readFixture(t *testing.T, path string) []byte {
	t.Helper()
	data, err := fs.ReadFile(fixtures.FS(), path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// repoOptions validates as inkt validate would inside the repository, with
// the template folder of ok-minima and an empty shared/.
func repoOptions(t *testing.T) contract.Options {
	t.Helper()
	return contract.Options{
		Origin:     contract.OriginRepository,
		RepoPolicy: true,
		Context: &contract.Context{
			TemplateDir: filepath.Join(fixturetest.DataDir(t), "templates", "prueba", "ok-minima"),
			FolderName:  "ok-minima",
			ParentName:  "prueba",
			SharedDir:   t.TempDir(),
			LicensesDir: filepath.Join(fixturetest.RepoRoot(t), "LICENSES"),
		},
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

func TestOkMinimaIsValid(t *testing.T) {
	m, rep := contract.ValidateManifest(readFixture(t, okMinima), repoOptions(t))
	if m == nil || rep.HasErrors() {
		t.Fatalf("ok-minima debe ser válida: %v", rep.Findings)
	}
	// The only warning: "prueba" is not a recommended category.
	for _, f := range rep.Warnings() {
		if f.Code != "V06" || f.Field != "category" {
			t.Errorf("aviso inesperado: %v", f)
		}
	}
	if m.ID != "ok-minima" || len(m.Parameters) != 7 || !m.Parameters[0].UsesProjectName() {
		t.Errorf("manifiesto mal leído: %+v", m)
	}
	if got := m.Parameters[1].EffectiveMaxLength(); got != contract.DefaultStringMaxLength {
		t.Errorf("maxLength por defecto = %d", got)
	}
	if got := m.Parameters[6].EffectiveMaxItems(); got != 5 {
		t.Errorf("maxItems de comite = %d, se esperaba 5", got)
	}
}

type manifestCase struct {
	Code  string         `json:"code"`
	Level contract.Level `json:"level"`
	Name  string         `json:"name"`
	Patch map[string]any `json:"patch"`
	Raw   *string        `json:"raw"`
}

// TestInvalidManifests runs every case of invalid/manifests.json: each must
// produce its finding and no other beyond those ok-minima already has.
func TestInvalidManifests(t *testing.T) {
	var file struct {
		Cases []manifestCase `json:"cases"`
	}
	fixturetest.ReadJSON(t, "invalid/manifests.json", &file)
	baseData := readFixture(t, okMinima)
	_, baseReport := contract.ValidateManifest(baseData, repoOptions(t))
	base := keysOf(baseReport)

	covered, names := map[string]bool{}, map[string]bool{}
	for _, c := range file.Cases {
		if names[c.Name] {
			t.Errorf("nombre de caso repetido: %s", c.Name)
		}
		names[c.Name] = true
		covered[c.Code] = true
		t.Run(c.Code+"/"+c.Name, func(t *testing.T) {
			data := caseData(t, baseData, c)
			_, rep := contract.ValidateManifest(data, repoOptions(t))
			want := findingKey{c.Code, c.Level}
			got := keysOf(rep)
			for k := range base {
				delete(got, k)
			}
			if !got[want] {
				t.Errorf("falta %s %s; hallazgos: %v", c.Code, c.Level, rep.Findings)
			}
			for k := range got {
				if k != want {
					t.Errorf("hallazgo inesperado %s %s; hallazgos: %v", k.code, k.level, rep.Findings)
				}
			}
		})
	}
	for n := 1; n <= 20; n++ {
		if code := fmt.Sprintf("V%02d", n); !covered[code] {
			t.Errorf("la regla %s no tiene ningún caso en invalid/manifests.json", code)
		}
	}
}

func caseData(t *testing.T, base []byte, c manifestCase) []byte {
	t.Helper()
	if c.Raw != nil {
		return []byte(*c.Raw)
	}
	dec := json.NewDecoder(bytes.NewReader(base))
	dec.UseNumber()
	var doc any
	if err := dec.Decode(&doc); err != nil {
		t.Fatal(err)
	}
	out, err := json.MarshalIndent(fixturetest.MergePatch(doc, c.Patch), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestFindingLinePointsAtTheField(t *testing.T) {
	data := caseData(t, readFixture(t, okMinima), manifestCase{Patch: map[string]any{
		"parameters": []any{map[string]any{"key": "titulo", "macro": "title", "type": "string", "label": "Título"}},
	}})
	_, rep := contract.ValidateManifest(data, repoOptions(t))
	lines := strings.Split(string(data), "\n")
	for _, f := range rep.Findings {
		if f.Code == "V12" {
			if f.Line < 1 || !strings.Contains(lines[f.Line-1], `"macro": "title"`) {
				t.Fatalf("línea %d = %q; se esperaba la del campo macro", f.Line, lines[max(f.Line-1, 0)])
			}
			return
		}
	}
	t.Fatalf("no se reportó V12: %v", rep.Findings)
}

func TestSyntaxErrorLine(t *testing.T) {
	_, rep := contract.ValidateManifest([]byte("{\n  \"schemaVersion\": 1,\n  \"id\": \n}"), contract.Options{})
	if len(rep.Findings) != 1 || rep.Findings[0].Code != "V01" || rep.Findings[0].Line != 4 {
		t.Fatalf("se esperaba un V01 en la línea 4: %v", rep.Findings)
	}
}

// TestFutureSchema: a client of contract v1 stops at V02 without judging the
// rest, and the template is incompatible, not broken.
func TestFutureSchema(t *testing.T) {
	m, rep := contract.ValidateManifest(readFixture(t, "templates/prueba/futura/template.json"), contract.Options{Origin: contract.OriginPublished})
	if m != nil || len(rep.Findings) != 1 || rep.Findings[0].Code != "V02" {
		t.Fatalf("se esperaba solo V02: %v", rep.Findings)
	}
}

func TestPublishedManifestNeedsDist(t *testing.T) {
	base := readFixture(t, okMinima)
	opt := contract.Options{Origin: contract.OriginPublished}

	_, rep := contract.ValidateManifest(base, opt)
	if !keysOf(rep)[findingKey{"V19", contract.LevelError}] {
		t.Errorf("sin dist debe dar V19: %v", rep.Findings)
	}

	withDist := caseData(t, base, manifestCase{Patch: map[string]any{"dist": map[string]any{
		"url":    "https://github.com/galigaribaldi/Memorandum-tex/releases/download/ok-minima-v1.0.0/ok-minima-1.0.0.zip",
		"sha256": strings.Repeat("ab", 32),
		"size":   json.Number("1234"),
	}}})
	if _, rep := contract.ValidateManifest(withDist, opt); rep.HasErrors() {
		t.Errorf("un manifiesto publicado con dist debe ser válido: %v", rep.Findings)
	}

	badDist := caseData(t, base, manifestCase{Patch: map[string]any{"dist": map[string]any{
		"url": "http://x", "sha256": "ABC", "size": json.Number("0"),
	}}})
	_, rep = contract.ValidateManifest(badDist, opt)
	if n := len(rep.Errors()); n != 3 {
		t.Errorf("dist inválido debe dar 3 errores (url, sha256, size), dio %d: %v", n, rep.Findings)
	}
}

// TestPublishedPreviewPDF: the sample PDF of §6.1 is optional when
// published, and when present it is checked by V17.
func TestPublishedPreviewPDF(t *testing.T) {
	dist := map[string]any{
		"url":    "https://github.com/galigaribaldi/Memorandum-tex/releases/download/ok-minima-v1.0.0/ok-minima-1.0.0.zip",
		"sha256": strings.Repeat("ab", 32), "size": json.Number("1234"),
	}
	opt := contract.Options{Origin: contract.OriginPublished}
	base := readFixture(t, okMinima)

	good := caseData(t, base, manifestCase{Patch: map[string]any{"dist": dist, "previewPdf": map[string]any{
		"url":    "https://github.com/galigaribaldi/Memorandum-tex/releases/download/ok-minima-v1.0.0/ok-minima-1.0.0-preview.pdf",
		"sha256": strings.Repeat("cd", 32), "size": json.Number("81234"), "pages": json.Number("3"),
	}}})
	m, rep := contract.ValidateManifest(good, opt)
	if rep.HasErrors() || m.PreviewPDF == nil || m.PreviewPDF.Pages != 3 {
		t.Fatalf("una vista previa válida debe aceptarse: %v", rep.Findings)
	}

	bad := caseData(t, base, manifestCase{Patch: map[string]any{"dist": dist, "previewPdf": map[string]any{
		"url": "http://x", "sha256": "ABC", "size": json.Number("5242881"), "pages": json.Number("0"),
	}}})
	_, rep = contract.ValidateManifest(bad, opt)
	errs := rep.Errors()
	if len(errs) != 4 {
		t.Fatalf("se esperaban 4 errores V17 (url, sha256, size > 5 MB, pages), hubo %d: %v", len(errs), rep.Findings)
	}
	for _, f := range errs {
		if f.Code != "V17" {
			t.Errorf("se esperaba V17: %v", f)
		}
	}

	typo := caseData(t, base, manifestCase{Patch: map[string]any{"dist": dist, "previewPdf": map[string]any{
		"url": "https://x/p.pdf", "sha256": strings.Repeat("cd", 32), "size": json.Number("1"), "pages": json.Number("1"),
		"paginas": json.Number("1"),
	}}})
	if _, rep := contract.ValidateManifest(typo, opt); !keysOf(rep)[findingKey{"V20", contract.LevelWarning}] {
		t.Errorf("un campo desconocido dentro de previewPdf debe avisar: %v", rep.Findings)
	}
}

// TestThirdPartyCatalogWithoutPolicy: without RepoPolicy, English and lualatex
// only are acceptable (another catalog, other rules).
func TestThirdPartyCatalogWithoutPolicy(t *testing.T) {
	data := caseData(t, readFixture(t, okMinima), manifestCase{Patch: map[string]any{
		"language": "en-US",
		"license":  "MIT",
		"build":    map[string]any{"engines": []any{"lualatex"}, "defaultEngine": "lualatex", "bibliography": "biber"},
	}})
	if _, rep := contract.ValidateManifest(data, contract.Options{}); rep.HasErrors() {
		t.Errorf("sin la política del repositorio debe ser válido: %v", rep.Findings)
	}
}
