package params_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/galigaribaldi/Memorandum-tex/tools/fixtures"
	"github.com/galigaribaldi/Memorandum-tex/tools/internal/fixturetest"
	"github.com/galigaribaldi/Memorandum-tex/tools/pkg/contract"
	"github.com/galigaribaldi/Memorandum-tex/tools/pkg/params"
)

func TestEscapeTeX(t *testing.T) {
	var file struct {
		Cases []struct{ In, Out string } `json:"cases"`
	}
	fixturetest.ReadJSON(t, "params/escape-cases.json", &file)
	if len(file.Cases) == 0 {
		t.Fatal("escape-cases.json no tiene casos")
	}
	for _, c := range file.Cases {
		if got := params.EscapeTeX(c.In); got != c.Out {
			t.Errorf("EscapeTeX(%q) = %q, se esperaba %q", c.In, got, c.Out)
		}
	}
}

type resolveCase struct {
	Name        string            `json:"name"`
	ProjectName string            `json:"projectName"`
	Now         string            `json:"now"`
	Values      map[string]string `json:"values"`
	Expect      struct {
		Values    map[string]any `json:"values"`
		InkParams string         `json:"inkParams"`
		Errors    []struct {
			Key  string `json:"key"`
			Code string `json:"code"`
		} `json:"errors"`
	} `json:"expect"`
}

// TestResolveCases runs params/resolve-cases.json and compares ink-params.tex
// byte for byte with the golden files.
func TestResolveCases(t *testing.T) {
	var file struct {
		Manifest string        `json:"manifest"`
		Cases    []resolveCase `json:"cases"`
	}
	fixturetest.ReadJSON(t, "params/resolve-cases.json", &file)
	manifestPath := path.Join("params", file.Manifest) // relative to the case file
	data, err := fs.ReadFile(fixtures.FS(), path.Clean(manifestPath))
	if err != nil {
		t.Fatal(err)
	}
	m, rep := contract.ValidateManifest(data, contract.Options{})
	if m == nil || rep.HasErrors() {
		t.Fatalf("manifiesto inválido: %v", rep.Findings)
	}

	goldens := 0
	for _, c := range file.Cases {
		t.Run(c.Name, func(t *testing.T) {
			now, err := time.Parse("2006-01-02", c.Now)
			if err != nil {
				t.Fatal(err)
			}
			res, err := params.Resolve(m, params.Input{Values: c.Values, ProjectName: c.ProjectName, Now: now})

			if c.Expect.Errors != nil {
				var perr *params.Error
				if !errors.As(err, &perr) {
					t.Fatalf("se esperaban errores, dio %v (%v)", err, res)
				}
				got := make([][2]string, len(perr.Fields))
				for i, f := range perr.Fields {
					got[i] = [2]string{f.Key, f.Code}
				}
				want := make([][2]string, len(c.Expect.Errors))
				for i, e := range c.Expect.Errors {
					want[i] = [2]string{e.Key, e.Code}
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("errores = %v, se esperaban %v", got, want)
				}
				return
			}

			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if gotJSON, wantJSON := mustJSON(t, res.Map()), mustJSON(t, c.Expect.Values); !bytes.Equal(gotJSON, wantJSON) {
				t.Errorf("valores = %s\nse esperaban %s", gotJSON, wantJSON)
			}
			if c.Expect.InkParams != "" {
				goldens++
				want, err := fs.ReadFile(fixtures.FS(), path.Join("params", c.Expect.InkParams))
				if err != nil {
					t.Fatal(err)
				}
				if got := params.RenderInkParams(res); !bytes.Equal(got, want) {
					t.Errorf("ink-params.tex distinto del esperado:\n--- obtenido\n%s--- esperado\n%s", got, want)
				}
			}
		})
	}
	if goldens != 2 {
		t.Errorf("se compararon %d archivos ink-params.tex esperados; hay 2 en params/", goldens)
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestResolveUsesTodayAndOmitsOptional(t *testing.T) {
	var m contract.Manifest
	fixturetest.ReadJSON(t, "templates/prueba/ok-minima/template.json", &m)
	now := time.Date(2030, 1, 2, 23, 0, 0, 0, time.Local)
	res, err := params.Resolve(&m, params.Input{Values: map[string]string{"autor": "X"}, ProjectName: "P", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	if it, _ := res.Get("fecha"); it.String != "2030-01-02" {
		t.Errorf("today = %q, se esperaba la fecha local 2030-01-02", it.String)
	}
	if _, ok := res.Get("nota"); ok {
		t.Error("un parámetro opcional sin valor no debe resolverse")
	}
	if bytes.Contains(params.RenderInkParams(res), []byte(`\Nota`)) {
		t.Error("un parámetro sin valor no debe escribirse en ink-params.tex")
	}
	res, err = params.Resolve(&m, params.Input{Values: map[string]string{"autor": "X", "fecha": "today"}, ProjectName: "P", Now: now})
	if it, _ := res.Get("fecha"); err != nil || it.String != "2030-01-02" {
		t.Errorf("fecha=today = %q, %v", it.String, err)
	}
}

func TestParseAssignments(t *testing.T) {
	got, err := params.ParseAssignments([]string{"portada=false", "titulo=Uno = dos", "nota="})
	want := map[string]string{"portada": "false", "titulo": "Uno = dos", "nota": ""}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("ParseAssignments = %v, %v", got, err)
	}
	for _, bad := range [][]string{{"portada"}, {"=valor"}, {"a=1", "a=2"}} {
		if _, err := params.ParseAssignments(bad); err == nil {
			t.Errorf("ParseAssignments(%q) debe fallar", bad)
		}
	}
}
