// Package compat tells whether a template can be used on this machine
// (docs/integracion-creative-ink.md §7). It never runs processes: the caller
// (Creative-Ink) says which engines are installed and passes a function that
// asks kpsewhich for files.
package compat

import (
	"slices"
	"strings"

	"github.com/galigaribaldi/Memorandum-tex/tools/pkg/contract"
)

// Status of a template on this machine.
type Status string

// Statuses, from the most to the least severe. Evaluate returns the most
// severe that applies.
const (
	// Incompatible: another schemaVersion, an invalid published manifest or a
	// newer Creative-Ink required. Update Creative-Ink (or report the catalog).
	Incompatible Status = "incompatible"
	// Blocked: the template asks for --shell-escape, which v1 never allows,
	// not even trusting the folder (contract §3.1, MAP decision 15).
	Blocked Status = "blocked"
	// MissingEngine: none of build.engines is installed.
	MissingEngine Status = "missing-engine"
	// MissingPackages: some requires.files is not found by kpsewhich.
	MissingPackages Status = "missing-packages"
	// Ready: it can be downloaded and compiled.
	Ready Status = "ready"
)

var severity = map[Status]int{Incompatible: 4, Blocked: 3, MissingEngine: 2, MissingPackages: 1, Ready: 0}

// Env describes this machine. Zero fields skip their check.
type Env struct {
	// Engines installed: "pdflatex", "xelatex", "lualatex". Nil skips the check.
	Engines map[string]bool
	// HasFiles answers, for each name (natbib.sty, spanish.ldf), whether TeX
	// finds it. Creative-Ink implements it with one kpsewhich call. Nil skips
	// the check.
	HasFiles func(names []string) map[string]bool
	// InkVersion is the running Creative-Ink version (semver). "" skips the
	// requires.creativeInk check.
	InkVersion string
}

// Report is the verdict for one template.
type Report struct {
	Status Status `json:"status"`
	// Reasons explain every problem found, in the order they were checked, in Spanish.
	Reasons []string `json:"reasons,omitempty"`
	// MissingEngines lists build.engines that are not installed (all of them
	// when Status is MissingEngine).
	MissingEngines []string `json:"missingEngines,omitempty"`
	// MissingFiles lists requires.files that kpsewhich did not find.
	MissingFiles []string `json:"missingFiles,omitempty"`
	// Hints are commands for the USER to run (never executed by the library
	// nor by Creative-Ink), e.g. "tlmgr install natbib".
	Hints []string `json:"hints,omitempty"`
}

func (r *Report) raise(s Status, reason string) {
	if severity[s] > severity[r.Status] {
		r.Status = s
	}
	r.Reasons = append(r.Reasons, reason)
}

// Evaluate judges a parsed manifest. For raw JSON from a catalog use
// EvaluateJSON, which also handles unknown schema versions.
func Evaluate(m *contract.Manifest, env Env) Report {
	r := Report{Status: Ready}
	if m.SchemaVersion != contract.SchemaVersion {
		r.raise(Incompatible, "usa una versión del contrato que esta versión de Creative-Ink no conoce")
	}
	if m.Requires != nil && m.Requires.CreativeInk != "" && env.InkVersion != "" {
		need, okNeed := contract.MinimumCreativeInk(m.Requires.CreativeInk)
		have, okHave := contract.ParseSemver(env.InkVersion)
		if okNeed && okHave && have.Compare(need) < 0 {
			r.raise(Incompatible, "necesita Creative-Ink "+need.String()+" o posterior (tienes "+have.String()+")")
		}
	}
	if m.Build.ShellEscape {
		r.raise(Blocked, "pide --shell-escape (ejecutar comandos del sistema), que no está permitido")
	}
	if env.Engines != nil {
		for _, e := range m.Build.Engines {
			if !env.Engines[e] {
				r.MissingEngines = append(r.MissingEngines, e)
			}
		}
		if len(m.Build.Engines) > 0 && len(r.MissingEngines) == len(m.Build.Engines) {
			r.raise(MissingEngine, "no está instalado ninguno de sus motores: "+strings.Join(m.Build.Engines, ", "))
			r.Hints = append(r.Hints, "Instala una distribución de TeX con "+m.Build.DefaultEngine+" (TeX Live, MacTeX o MiKTeX)")
		}
	}
	if m.Requires != nil && len(m.Requires.Files) > 0 && env.HasFiles != nil {
		found := env.HasFiles(slices.Clone(m.Requires.Files))
		for _, f := range m.Requires.Files {
			if !found[f] {
				r.MissingFiles = append(r.MissingFiles, f)
			}
		}
		if len(r.MissingFiles) > 0 {
			r.raise(MissingPackages, "TeX no encuentra: "+strings.Join(r.MissingFiles, ", "))
			if len(m.Requires.Packages) > 0 {
				r.Hints = append(r.Hints, "tlmgr install "+strings.Join(m.Requires.Packages, " "))
			}
		}
	}
	return r
}

// EvaluateJSON validates a published manifest (without this repository's
// policy) and judges it. An unknown schemaVersion or an invalid manifest is
// Incompatible: a result, never an error that hides the rest of the catalog.
func EvaluateJSON(data []byte, env Env) (Report, *contract.Manifest) {
	m, rep := contract.ValidateManifest(data, contract.Options{Origin: contract.OriginPublished})
	if m == nil || rep.HasErrors() {
		r := Report{Status: Incompatible}
		for _, f := range rep.Errors() {
			r.Reasons = append(r.Reasons, "manifiesto inválido: "+f.String())
		}
		return r, m
	}
	return Evaluate(m, env), m
}

// UpdateAvailable reports whether latest is newer than installed. Invalid
// versions never offer an update.
func UpdateAvailable(installed, latest string) bool {
	have, ok1 := contract.ParseSemver(installed)
	next, ok2 := contract.ParseSemver(latest)
	return ok1 && ok2 && next.Compare(have) > 0
}
