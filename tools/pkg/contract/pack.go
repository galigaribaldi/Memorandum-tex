package contract

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

// PackOptions of ValidatePack.
type PackOptions struct {
	// FileName is the pack's file name without ".json" (V50); "" skips the check.
	FileName string
	// Known reports whether a template id exists (V52); nil skips the check.
	Known func(id string) bool
	// MissingIsError makes an unknown template an error (inkt index) instead
	// of a warning (inkt validate), §8.
	MissingIsError bool
	// AllowReserved accepts the id "todo"; only inkt index generates it.
	AllowReserved bool
}

var knownPackFields = []string{"$schema", "schemaVersion", "id", "name", "description", "templates"}

// ValidatePack parses and validates packs/<id>.json (§8, rules V50–V52).
func ValidatePack(data []byte, opt PackOptions) (*Pack, Report) {
	r := newReporter(data)
	raw, ok := checkJSON(r, data, MaxManifestBytes)
	if !ok {
		return nil, r.report
	}
	if !checkSchemaVersion(r, raw) {
		return nil, r.report
	}
	var p Pack
	if err := json.Unmarshal(data, &p); err != nil {
		r.errorf("V51", "", "no se pudo leer el paquete: %v", err)
		return nil, r.report
	}
	warnUnknown(r, "", raw, knownPackFields)

	switch {
	case !ValidID(p.ID):
		r.errorf("V50", "id", "%q no es un id válido: kebab-case ASCII de 2 a 40 caracteres", p.ID)
	case p.ID == ReservedPackID && !opt.AllowReserved:
		r.errorf("V50", "id", "%q está reservado: lo genera inkt index con todas las plantillas", ReservedPackID)
	case opt.FileName != "" && p.ID != opt.FileName:
		r.errorf("V50", "id", "es %q pero el archivo se llama %s.json", p.ID, opt.FileName)
	}
	if n := utf8.RuneCountInString(p.Name); n < 1 || n > 80 {
		r.errorf("V51", "name", "debe tener de 1 a 80 caracteres (tiene %d)", n)
	}
	if n := utf8.RuneCountInString(p.Description); n < 1 || n > 300 || strings.ContainsAny(p.Description, "\r\n") {
		r.errorf("V51", "description", "debe ser una línea de 1 a 300 caracteres")
	}
	if len(p.Templates) < 1 || len(p.Templates) > 100 {
		r.errorf("V51", "templates", "debe tener de 1 a 100 plantillas (tiene %d)", len(p.Templates))
	}
	seen := map[string]bool{}
	for i, t := range p.Templates {
		f := fmt.Sprintf("templates[%d]", i)
		switch {
		case !ValidID(t.ID):
			r.errorf("V51", f+".id", "%q no es un id válido", t.ID)
			continue
		case seen[t.ID]:
			r.errorf("V51", f+".id", "%q está repetida", t.ID)
		}
		seen[t.ID] = true
		if t.Version != "" && !ValidSemver(t.Version) {
			r.errorf("V51", f+".version", "%q no es semver", t.Version)
		}
		if opt.Known != nil && !opt.Known(t.ID) {
			level := LevelWarning
			if opt.MissingIsError {
				level = LevelError
			}
			r.add("V52", level, f+".id", "la plantilla %q todavía no existe", t.ID)
		}
	}
	return &p, r.report
}
