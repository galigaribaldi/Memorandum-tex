// Package params resolves the parameters of a template (docs/contrato-plantillas.md
// §4) from what the user typed (clave=valor) and writes ink-params.tex (§10).
//
// It is pure: no files, no network, no processes. The caller (inkt or
// Creative-Ink) decides where ink-params.tex goes.
package params

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/galigaribaldi/Memorandum-tex/tools/pkg/contract"
)

// Error codes of a FieldError (§4.3). They are stable: the tests of both
// repositories compare them.
const (
	CodeUnknown  = "unknown"
	CodeRequired = "required"
	CodeType     = "type"
	CodeEnum     = "enum"
	CodeLength   = "length"
	CodeItems    = "items"
)

// FieldError is why one parameter could not be resolved.
type FieldError struct {
	Key     string `json:"key"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error lists every FieldError: first the manifest's parameters in order
// (at most one each), then unknown keys in byte order.
type Error struct {
	Fields []FieldError `json:"fields"`
}

func (e *Error) Error() string {
	parts := make([]string, len(e.Fields))
	for i, f := range e.Fields {
		parts[i] = f.Key + ": " + f.Message
	}
	return "parámetros inválidos: " + strings.Join(parts, "; ")
}

// Input is what Resolve needs besides the manifest.
type Input struct {
	// Values are the user's clave=valor, as typed.
	Values map[string]string
	// ProjectName fills the fromProjectName parameter when the user gave no value.
	ProjectName string
	// Now resolves "today"; the zero value means time.Now().
	Now time.Time
}

// Item is the resolved value of one parameter. String holds string, enum and
// date values; Bool booleans; List lists.
type Item struct {
	Param  contract.Parameter
	String string
	Bool   bool
	List   []string
}

// Value returns the item as a JSON-friendly value: string, bool or []string.
func (it Item) Value() any {
	switch it.Param.Type {
	case contract.TypeBoolean:
		return it.Bool
	case contract.TypeList:
		return it.List
	default:
		return it.String
	}
}

// Resolved holds the values of a manifest's parameters, in manifest order.
// Optional parameters without a value are absent.
type Resolved struct {
	ID, Version string
	Items       []Item
}

// Get returns the item of key.
func (r *Resolved) Get(key string) (Item, bool) {
	for _, it := range r.Items {
		if it.Param.Key == key {
			return it, true
		}
	}
	return Item{}, false
}

// Map returns key → Value() for every resolved parameter.
func (r *Resolved) Map() map[string]any {
	out := make(map[string]any, len(r.Items))
	for _, it := range r.Items {
		out[it.Param.Key] = it.Value()
	}
	return out
}

// Resolve applies, for each parameter: the user's value, then
// fromProjectName, then default; a required parameter left without a value
// is an error (§4.3). The error, if any, is an *Error.
func Resolve(m *contract.Manifest, in Input) (*Resolved, error) {
	now := in.Now
	if now.IsZero() {
		now = time.Now()
	}
	res := &Resolved{ID: m.ID, Version: m.Version}
	var errs []FieldError
	fail := func(key, code, format string, args ...any) {
		errs = append(errs, FieldError{Key: key, Code: code, Message: fmt.Sprintf(format, args...)})
	}

	known := make(map[string]bool, len(m.Parameters))
	for _, p := range m.Parameters {
		known[p.Key] = true
		text, given := in.Values[p.Key]
		if !given && p.UsesProjectName() && in.ProjectName != "" {
			text, given = in.ProjectName, true
		}
		var item Item
		var code, msg string
		switch {
		case given:
			item, code, msg = parseText(p, text, now)
		case p.HasDefault():
			item, code, msg = parseDefault(p, now)
		case p.Required:
			code, msg = CodeRequired, "es obligatorio y no tiene valor por defecto"
		default:
			continue // optional, no value: not written to ink-params.tex
		}
		if code != "" {
			fail(p.Key, code, "%s", msg)
			continue
		}
		res.Items = append(res.Items, item)
	}

	var unknown []string
	for k := range in.Values {
		if !known[k] {
			unknown = append(unknown, k)
		}
	}
	sort.Strings(unknown)
	for _, k := range unknown {
		fail(k, CodeUnknown, "la plantilla %s no tiene el parámetro %q", m.ID, k)
	}
	if len(errs) > 0 {
		return nil, &Error{Fields: errs}
	}
	return res, nil
}

// parseText interprets what the user typed. It returns a code and message
// when the text is not valid for the parameter.
func parseText(p contract.Parameter, text string, now time.Time) (Item, string, string) {
	it := Item{Param: p}
	switch p.Type {
	case contract.TypeString:
		if code, msg := checkLine(text, p.EffectiveMaxLength(), CodeLength); code != "" {
			return it, code, msg
		}
		it.String = text
	case contract.TypeBoolean:
		switch strings.ToLower(strings.TrimSpace(text)) {
		case "true", "si", "sí":
			it.Bool = true
		case "false", "no":
			it.Bool = false
		default:
			return it, CodeType, fmt.Sprintf("%q no es un booleano (true, false, sí, no)", text)
		}
	case contract.TypeEnum:
		if !slices.Contains(p.Values, text) {
			return it, CodeEnum, fmt.Sprintf("%q no es uno de: %s", text, strings.Join(p.Values, ", "))
		}
		it.String = text
	case contract.TypeDate:
		switch {
		case text == contract.DefaultToday:
			it.String = now.Format("2006-01-02")
		case contract.ValidDate(text):
			it.String = text
		default:
			return it, CodeType, fmt.Sprintf("%q no es una fecha real AAAA-MM-DD", text)
		}
	case contract.TypeList:
		items := []string{}
		for _, s := range strings.Split(text, ";") {
			if s = strings.TrimSpace(s); s != "" {
				items = append(items, s)
			}
		}
		if code, msg := checkList(p, items); code != "" {
			return it, code, msg
		}
		it.List = items
	default:
		return it, CodeType, fmt.Sprintf("tipo desconocido %q", p.Type)
	}
	return it, "", ""
}

// parseDefault decodes the manifest's default (already checked by V14).
func parseDefault(p contract.Parameter, now time.Time) (Item, string, string) {
	it := Item{Param: p}
	var err error
	switch p.Type {
	case contract.TypeBoolean:
		err = json.Unmarshal(p.Default, &it.Bool)
	case contract.TypeList:
		err = json.Unmarshal(p.Default, &it.List)
		if it.List == nil {
			it.List = []string{}
		}
	default:
		err = json.Unmarshal(p.Default, &it.String)
		if p.Type == contract.TypeDate && it.String == contract.DefaultToday {
			it.String = now.Format("2006-01-02")
		}
	}
	if err != nil {
		return it, CodeType, "el valor por defecto del manifiesto no es válido (revisa V14)"
	}
	return it, "", ""
}

func checkLine(s string, max int, code string) (string, string) {
	if strings.ContainsAny(s, "\r\n") {
		return code, "debe ser una sola línea"
	}
	if n := utf8.RuneCountInString(s); n > max {
		return code, fmt.Sprintf("tiene %d caracteres; el máximo es %d", n, max)
	}
	return "", ""
}

func checkList(p contract.Parameter, items []string) (string, string) {
	if len(items) > p.EffectiveMaxItems() {
		return CodeItems, fmt.Sprintf("tiene %d elementos; el máximo es %d", len(items), p.EffectiveMaxItems())
	}
	for i, s := range items {
		if code, msg := checkLine(s, p.EffectiveMaxLength(), CodeItems); code != "" {
			return code, fmt.Sprintf("elemento %d: %s", i+1, msg)
		}
	}
	return "", ""
}
