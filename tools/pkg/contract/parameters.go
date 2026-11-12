package contract

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

// parameters applies V12–V15 (§4).
func (v *manifestValidator) parameters() {
	r := v.r
	if len(v.m.Parameters) > MaxParameters {
		r.errorf("V13", "parameters", "hay %d parámetros; el máximo es %d", len(v.m.Parameters), MaxParameters)
	}
	keys, macros := map[string]bool{}, map[string]bool{}
	fromProject := 0
	for i, p := range v.m.Parameters {
		f := fmt.Sprintf("parameters[%d]", i)
		checkParameterIdentity(r, f, p, keys, macros)
		if !checkParameterType(r, f, p) {
			continue // unknown type: its default cannot be judged
		}
		if p.UsesProjectName() {
			fromProject++
			if p.Type != TypeString {
				r.errorf("V15", f+".fromProjectName", "solo vale para parámetros de tipo string")
			}
		}
		checkParameterDefault(r, f, p)
	}
	if fromProject > 1 {
		r.errorf("V15", "parameters", "fromProjectName aparece en %d parámetros; solo puede haber uno", fromProject)
	}
}

// checkParameterIdentity applies V12: key and macro.
func checkParameterIdentity(r *reporter, f string, p Parameter, keys, macros map[string]bool) {
	switch {
	case !reKey.MatchString(p.Key):
		r.errorf("V12", f+".key", "%q no es válida: camelCase ASCII que empieza en minúscula, hasta 40 caracteres", p.Key)
	case keys[p.Key]:
		r.errorf("V12", f+".key", "%q está repetida", p.Key)
	}
	keys[p.Key] = true
	switch {
	case !reMacro.MatchString(p.Macro):
		r.errorf("V12", f+".macro", "%q no es válida: solo letras ASCII, hasta 40", p.Macro)
	case IsReservedMacro(p.Macro):
		r.errorf("V12", f+".macro", "\\%s está reservada (LaTeX, Creative-Ink o prefijo ink)", p.Macro)
	case macros[p.Macro]:
		r.errorf("V12", f+".macro", "\\%s está repetida", p.Macro)
	}
	macros[p.Macro] = true
}

// checkParameterType applies V13: known type, labels and the fields each
// type allows. It returns false for an unknown type.
func checkParameterType(r *reporter, f string, p Parameter) bool {
	if n := utf8.RuneCountInString(p.Label); n < 1 || n > 60 {
		r.errorf("V13", f+".label", "debe tener de 1 a 60 caracteres (tiene %d)", n)
	}
	if utf8.RuneCountInString(p.Description) > 300 {
		r.errorf("V13", f+".description", "supera los 300 caracteres")
	}
	if utf8.RuneCountInString(p.Group) > 40 {
		r.errorf("V13", f+".group", "supera los 40 caracteres")
	}
	forbid := func(present bool, field string) {
		if present {
			r.errorf("V13", f+"."+field, "no se usa con el tipo %s", p.Type)
		}
	}
	inRange := func(n *int, lo, hi int, field string) {
		if n != nil && (*n < lo || *n > hi) {
			r.errorf("V13", f+"."+field, "%d no está entre %d y %d", *n, lo, hi)
		}
	}
	switch p.Type {
	case TypeString:
		forbid(p.Values != nil, "values")
		forbid(p.MaxItems != nil, "maxItems")
		inRange(p.MaxLength, 1, 1000, "maxLength")
	case TypeBoolean, TypeDate:
		forbid(p.Values != nil, "values")
		forbid(p.MaxLength != nil, "maxLength")
		forbid(p.MaxItems != nil, "maxItems")
	case TypeEnum:
		forbid(p.MaxLength != nil, "maxLength")
		forbid(p.MaxItems != nil, "maxItems")
		if len(p.Values) < 1 || len(p.Values) > 50 {
			r.errorf("V13", f+".values", "debe tener de 1 a 50 valores (tiene %d)", len(p.Values))
		}
		seen := map[string]bool{}
		for j, val := range p.Values {
			vf := fmt.Sprintf("%s.values[%d]", f, j)
			switch {
			case !reEnumValue.MatchString(val):
				r.errorf("V13", vf, "%q no es válido: letras, dígitos y guiones, hasta 40 (se usa en nombres de archivo)", val)
			case seen[val]:
				r.errorf("V13", vf, "%q está repetido", val)
			}
			seen[val] = true
		}
	case TypeList:
		forbid(p.Values != nil, "values")
		inRange(p.MaxItems, 1, 50, "maxItems")
		inRange(p.MaxLength, 1, 1000, "maxLength")
	default:
		r.errorf("V13", f+".type", "%q no es un tipo (string, boolean, enum, date, list)", p.Type)
		return false
	}
	return true
}

// checkParameterDefault applies V14.
func checkParameterDefault(r *reporter, f string, p Parameter) {
	df := f + ".default"
	has := p.HasDefault()
	switch p.Type {
	case TypeString:
		if !has {
			return
		}
		var s string
		if json.Unmarshal(p.Default, &s) != nil {
			r.errorf("V14", df, "debe ser un texto")
			return
		}
		if msg := checkText(s, p.EffectiveMaxLength()); msg != "" {
			r.errorf("V14", df, "%s", msg)
		}
	case TypeBoolean:
		var b bool
		if !has || json.Unmarshal(p.Default, &b) != nil {
			r.errorf("V14", df, "un booleano necesita default true o false")
		}
	case TypeEnum:
		var s string
		if !has || json.Unmarshal(p.Default, &s) != nil {
			r.errorf("V14", df, "un enum necesita default, uno de sus valores")
			return
		}
		if !contains(p.Values, s) {
			r.errorf("V14", df, "%q no está entre los valores (%s)", s, strings.Join(p.Values, ", "))
		}
	case TypeDate:
		if !has {
			return
		}
		var s string
		if json.Unmarshal(p.Default, &s) != nil || (s != DefaultToday && !ValidDate(s)) {
			r.errorf("V14", df, "debe ser %q o una fecha real AAAA-MM-DD", DefaultToday)
		}
	case TypeList:
		if !has {
			return
		}
		var items []string
		if json.Unmarshal(p.Default, &items) != nil {
			r.errorf("V14", df, "debe ser una lista de textos")
			return
		}
		if len(items) > p.EffectiveMaxItems() {
			r.errorf("V14", df, "tiene %d elementos; el máximo es %d", len(items), p.EffectiveMaxItems())
		}
		for j, it := range items {
			if msg := checkText(it, p.EffectiveMaxLength()); msg != "" {
				r.errorf("V14", fmt.Sprintf("%s[%d]", df, j), "%s", msg)
			}
		}
	}
}

// checkText returns why s is not a valid single-line value of at most max
// characters, or "" if it is.
func checkText(s string, max int) string {
	if strings.ContainsAny(s, "\r\n") {
		return "debe ser una sola línea"
	}
	if n := utf8.RuneCountInString(s); n > max {
		return fmt.Sprintf("tiene %d caracteres; el máximo es %d", n, max)
	}
	return ""
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
