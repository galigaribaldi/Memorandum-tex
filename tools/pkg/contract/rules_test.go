package contract_test

import (
	"testing"

	"github.com/galigaribaldi/Memorandum-tex/tools/pkg/contract"
)

func TestValidRelPath(t *testing.T) {
	for p, want := range map[string]bool{
		"main.tex":                 true,
		"secciones/1-intro.tex":    true,
		"Latex/ink/ink-motor.tex":  true,
		"":                         false,
		"/etc/hosts":               false,
		"../fuera.tex":             false,
		"a/./b.tex":                false,
		"a//b.tex":                 false,
		"C:/fuera.tex":             false,
		`..\..\fuera.tex`:          false,
		"mi archivo.tex":           false,
		"introducción.tex":         false,
		"secciones/../../main.tex": false,
	} {
		if got := contract.ValidRelPath(p); got != want {
			t.Errorf("ValidRelPath(%q) = %v, se esperaba %v", p, got, want)
		}
	}
}

func TestParseLicense(t *testing.T) {
	for expr, want := range map[string]int{
		"MIT-0":                       1,
		"MIT-0 AND LicenseRef-isprs":  2,
		"MIT-0 OR MIT":                0,
		"MIT-0 AND":                   0,
		"(MIT-0)":                     0,
		"Apache-2.0 WITH LLVM-except": 0,
		"":                            0,
	} {
		ids, ok := contract.ParseLicense(expr)
		if (want > 0) != ok || len(ids) != want {
			t.Errorf("ParseLicense(%q) = %v, %v; se esperaban %d identificadores", expr, ids, ok, want)
		}
	}
}

func TestReservedMacros(t *testing.T) {
	for name, want := range map[string]bool{
		"title": true, "borrador": true, "SoloCapitulo": true, "inkItem": true, "InkColor": true,
		"TituloDoc": false, "Autor": false, "Titulo": false,
	} {
		if got := contract.IsReservedMacro(name); got != want {
			t.Errorf("IsReservedMacro(%q) = %v", name, got)
		}
	}
}

func TestSemver(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0}, {"1.0.0", "1.0.1", -1}, {"1.10.0", "1.9.9", 1}, {"2.0.0", "10.0.0", -1},
	}
	for _, c := range cases {
		a, okA := contract.ParseSemver(c.a)
		b, okB := contract.ParseSemver(c.b)
		if !okA || !okB || a.Compare(b) != c.want {
			t.Errorf("%s vs %s = %d, se esperaba %d", c.a, c.b, a.Compare(b), c.want)
		}
	}
	for _, bad := range []string{"1.0", "01.0.0", "1.0.0-beta", "v1.0.0", ""} {
		if _, ok := contract.ParseSemver(bad); ok {
			t.Errorf("ParseSemver(%q) debe fallar", bad)
		}
	}
	if min, ok := contract.MinimumCreativeInk(">=0.3.1"); !ok || min.String() != "0.3.1" {
		t.Errorf("MinimumCreativeInk = %v, %v", min, ok)
	}
	if _, ok := contract.MinimumCreativeInk("^1.0.0"); ok {
		t.Error("MinimumCreativeInk(^1.0.0) debe fallar")
	}
}

func TestValidDate(t *testing.T) {
	for s, want := range map[string]bool{"2026-11-10": true, "2024-02-29": true, "2026-02-30": false, "10/11/2026": false, "2026-1-1": false} {
		if got := contract.ValidDate(s); got != want {
			t.Errorf("ValidDate(%q) = %v", s, got)
		}
	}
}
