package compat_test

import (
	"io/fs"
	"slices"
	"strings"
	"testing"

	"github.com/galigaribaldi/Memorandum-tex/tools/fixtures"
	"github.com/galigaribaldi/Memorandum-tex/tools/pkg/compat"
	"github.com/galigaribaldi/Memorandum-tex/tools/pkg/contract"
)

var allEngines = map[string]bool{"pdflatex": true, "xelatex": true, "lualatex": true}

// kpsewhich simulates TeX: it finds every file except the fixture's invented package.
func kpsewhich(names []string) map[string]bool {
	out := map[string]bool{}
	for _, n := range names {
		out[n] = !strings.HasPrefix(n, "memorandum-paquete-inexistente")
	}
	return out
}

func manifest(t *testing.T, path string) *contract.Manifest {
	t.Helper()
	data, err := fs.ReadFile(fixtures.FS(), path)
	if err != nil {
		t.Fatal(err)
	}
	m, rep := contract.ValidateManifest(data, contract.Options{})
	if m == nil || rep.HasErrors() {
		t.Fatalf("%s: %v", path, rep.Findings)
	}
	return m
}

func TestFixtures(t *testing.T) {
	env := compat.Env{Engines: allEngines, HasFiles: kpsewhich, InkVersion: "0.1.0"}

	if r := compat.Evaluate(manifest(t, "templates/prueba/ok-minima/template.json"), env); r.Status != compat.Ready || len(r.Reasons) != 0 {
		t.Errorf("ok-minima: %+v", r)
	}

	r := compat.Evaluate(manifest(t, "templates/prueba/con-paquetes/template.json"), env)
	if r.Status != compat.MissingPackages ||
		!slices.Equal(r.MissingFiles, []string{"memorandum-paquete-inexistente.sty"}) ||
		!slices.Contains(r.Hints, "tlmgr install memorandum-paquete-inexistente") {
		t.Errorf("con-paquetes: %+v", r)
	}

	data, err := fs.ReadFile(fixtures.FS(), "templates/prueba/futura/template.json")
	if err != nil {
		t.Fatal(err)
	}
	if r, m := compat.EvaluateJSON(data, env); r.Status != compat.Incompatible || m != nil {
		t.Errorf("futura: %+v", r)
	}
}

// TestShellEscapeIsBlocked: even with everything installed, --shell-escape
// blocks the template, and it wins over a missing engine.
func TestShellEscapeIsBlocked(t *testing.T) {
	m := manifest(t, "templates/prueba/ok-minima/template.json")
	m.Build.ShellEscape = true
	if r := compat.Evaluate(m, compat.Env{Engines: allEngines, HasFiles: kpsewhich}); r.Status != compat.Blocked {
		t.Errorf("con todo instalado: %+v", r)
	}
	if r := compat.Evaluate(m, compat.Env{Engines: map[string]bool{}}); r.Status != compat.Blocked || len(r.Reasons) != 2 {
		t.Errorf("blocked debe ganar a missing-engine y reportar ambas razones: %+v", r)
	}
}

func TestMissingEngine(t *testing.T) {
	m := manifest(t, "templates/prueba/ok-minima/template.json")
	if r := compat.Evaluate(m, compat.Env{Engines: map[string]bool{"xelatex": true}}); r.Status != compat.Ready ||
		!slices.Equal(r.MissingEngines, []string{"pdflatex", "lualatex"}) {
		t.Errorf("con un motor basta: %+v", r)
	}
	if r := compat.Evaluate(m, compat.Env{Engines: map[string]bool{}}); r.Status != compat.MissingEngine || len(r.Hints) != 1 {
		t.Errorf("sin motores: %+v", r)
	}
}

func TestCreativeInkVersion(t *testing.T) {
	m := manifest(t, "templates/prueba/ok-minima/template.json") // requires >=0.1.0
	for version, want := range map[string]compat.Status{
		"0.0.9": compat.Incompatible, "0.1.0": compat.Ready, "1.4.2": compat.Ready, "": compat.Ready, "no-semver": compat.Ready,
	} {
		if r := compat.Evaluate(m, compat.Env{InkVersion: version}); r.Status != want {
			t.Errorf("Creative-Ink %q: %s, se esperaba %s (%v)", version, r.Status, want, r.Reasons)
		}
	}
}

func TestUnknownChecksAreSkipped(t *testing.T) {
	m := manifest(t, "templates/prueba/con-paquetes/template.json")
	if r := compat.Evaluate(m, compat.Env{}); r.Status != compat.Ready {
		t.Errorf("sin Env no se comprueba nada: %+v", r)
	}
}

func TestInvalidPublishedManifestIsIncompatible(t *testing.T) {
	r, _ := compat.EvaluateJSON([]byte(`{"schemaVersion":1,"id":"X"}`), compat.Env{})
	if r.Status != compat.Incompatible || len(r.Reasons) == 0 {
		t.Errorf("%+v", r)
	}
}

func TestUpdateAvailable(t *testing.T) {
	for _, c := range []struct {
		installed, latest string
		want              bool
	}{
		{"1.0.0", "1.0.1", true}, {"1.2.0", "1.10.0", true}, {"1.0.0", "1.0.0", false}, {"2.0.0", "1.9.9", false}, {"x", "1.0.0", false},
	} {
		if got := compat.UpdateAvailable(c.installed, c.latest); got != c.want {
			t.Errorf("UpdateAvailable(%s, %s) = %v", c.installed, c.latest, got)
		}
	}
}
