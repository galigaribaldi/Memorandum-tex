// Package validate checks a whole template folder of this repository
// (docs/contrato-plantillas.md §12): the manifest (V01–V20, through
// pkg/contract) and the folder's content (V30–V39). inkt validate and the CI
// use it; Creative-Ink does not need it (it checks what it downloads with
// bundle.SafeExtract).
package validate

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/galigaribaldi/Memorandum-tex/tools/pkg/contract"
)

// Options of Folder.
type Options struct {
	// SharedDir is the repository's shared/ (components, V18 and V34).
	SharedDir string
	// LicensesDir is the repository's LICENSES/ (V08).
	LicensesDir string
}

var (
	reChangelogVersion = func(v string) *regexp.Regexp {
		return regexp.MustCompile(`(?m)^##\s+` + regexp.QuoteMeta(v) + `(\s|$)`)
	}
	// V35: a path that leaves the template inside a command that loads a file.
	reOutsidePath = regexp.MustCompile(`\\(input|include|includegraphics|bibliography|graphicspath)\b.*(\.\./|\{\s*/)`)
	// V39: shell escape and what needs it, and the title markers of Creative-Ink.
	reShellEscape = regexp.MustCompile(`\\tikzexternalize|\\usetikzlibrary\{[^}]*\bexternal\b|\{minted\}|\\write18|\\ShellEscape`)
	reMarker      = regexp.MustCompile(`\{\{TIT(ULO|LE)\}\}`)
	reInkParams   = regexp.MustCompile(`\\InputIfFileExists\{ink-params(\.tex)?\}`)
	reProvide     = regexp.MustCompile(`\\providecommand\*?\{?\\([A-Za-z]+)`)
)

// Folder validates the template at dir (templates/<categoria>/<id>/). The
// report lists manifest findings first (Field is a JSON path) and then folder
// findings (Field is a path relative to dir, with "/").
func Folder(dir string, opt Options) contract.Report {
	var rep contract.Report
	add := func(code, field string, line int, format string, args ...any) {
		rep.Findings = append(rep.Findings, contract.Finding{
			Code: code, Level: contract.LevelError, Field: field, Line: line, Message: fmt.Sprintf(format, args...),
		})
	}

	data, err := os.ReadFile(filepath.Join(dir, "template.json"))
	if err != nil {
		add("V30", "template.json", 0, "falta el manifiesto")
		return rep
	}
	m, mrep := contract.ValidateManifest(data, contract.Options{
		Origin:     contract.OriginRepository,
		RepoPolicy: true,
		Context: &contract.Context{
			TemplateDir: dir,
			FolderName:  filepath.Base(dir),
			ParentName:  filepath.Base(filepath.Dir(dir)),
			SharedDir:   opt.SharedDir,
			LicensesDir: opt.LicensesDir,
		},
	})
	rep.Findings = append(rep.Findings, mrep.Findings...)
	if m == nil {
		return rep // without a manifest of this contract the folder rules cannot be judged
	}

	// V30: README, CHANGELOG with the version, src/.
	if !isFile(filepath.Join(dir, "README.md")) {
		add("V30", "README.md", 0, "falta el README de la plantilla")
	}
	if cl, err := os.ReadFile(filepath.Join(dir, "CHANGELOG.md")); err != nil {
		add("V30", "CHANGELOG.md", 0, "falta el CHANGELOG de la plantilla")
	} else if m.Version != "" && !reChangelogVersion(m.Version).Match(cl) {
		add("V30", "CHANGELOG.md", 0, "no tiene un encabezado «## %s» para la versión del manifiesto", m.Version)
	}
	src := filepath.Join(dir, "src")
	if info, err := os.Stat(src); err != nil || !info.IsDir() {
		add("V30", "src", 0, "falta la carpeta src/")
		return rep
	}

	s := scan{add: add, lower: map[string]string{}, macros: map[string]bool{}}
	s.walk(src, "src")
	for _, c := range m.Components { // V34 counts the components that inkt bundle will copy
		if opt.SharedDir != "" {
			s.count(filepath.Join(opt.SharedDir, filepath.FromSlash(c)))
		}
	}
	if s.files > contract.MaxBundleFiles {
		add("V34", "src", 0, "con sus componentes tiene %d archivos; el máximo es %d", s.files, contract.MaxBundleFiles)
	}
	if s.bytes > contract.MaxBundleBytes {
		add("V34", "src", 0, "con sus componentes ocupa %d MB; el máximo es %d MB", s.bytes>>20, contract.MaxBundleBytes>>20)
	}

	checkParameters(m, filepath.Join(src, filepath.FromSlash(m.Entry)), "src/"+m.Entry, s.macros, add)
	return rep
}

type scan struct {
	add          func(code, field string, line int, format string, args ...any)
	lower        map[string]string // lowercase path → first path seen (B10)
	macros       map[string]bool   // macros defined with \providecommand in src/
	files, bytes int64
}

// walk applies V31–V35, V38 and V39 to every entry under root.
func (s *scan) walk(root, prefix string) {
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		rel := prefix
		if r, _ := filepath.Rel(root, p); r != "." {
			rel = prefix + "/" + filepath.ToSlash(r)
		}
		if err != nil {
			s.add("V31", rel, 0, "no se pudo leer: %v", err)
			return nil
		}
		if rel == prefix {
			return nil
		}
		name := d.Name()
		if d.Type()&fs.ModeSymlink != 0 {
			s.add("V33", rel, 0, "es un enlace simbólico; copia el archivo en su lugar")
			return nil
		}
		if !contract.ValidRelPath(strings.TrimPrefix(rel, prefix+"/")) {
			s.add("V31", rel, 0, "nombre no portable: usa ASCII sin espacios ni acentos (letras, dígitos, . _ -)")
		}
		if first, dup := s.lower[strings.ToLower(rel)]; dup {
			s.add("V31", rel, 0, "solo difiere en mayúsculas de %s: en macOS y Windows serían el mismo archivo", first)
		} else {
			s.lower[strings.ToLower(rel)] = rel
		}
		if d.IsDir() {
			if strings.HasPrefix(name, ".") {
				s.add("V32", rel, 0, "carpeta oculta no permitida")
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			s.add("V33", rel, 0, "no es un archivo normal")
			return nil
		}
		if !contract.AllowedSourceFile(name) {
			s.add("V32", rel, 0, "archivo no permitido en src/ (%s)", whyForbidden(name))
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		s.files++
		s.bytes += info.Size()
		if contract.IsTextFile(name) {
			s.text(p, rel, name)
		}
		return nil
	})
}

// count adds the files of a component to the V34 totals.
func (s *scan) count(dir string) {
	_ = filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err == nil && d.Type().IsRegular() {
			if info, err := d.Info(); err == nil {
				s.files++
				s.bytes += info.Size()
			}
		}
		return nil
	})
}

// text applies V38 to every text file and V35, V39 and the \providecommand
// collection (V37) to LaTeX sources.
func (s *scan) text(p, rel, name string) {
	data, err := os.ReadFile(p)
	if err != nil {
		return
	}
	switch {
	case bytes.HasPrefix(data, []byte("\xef\xbb\xbf")):
		s.add("V38", rel, 1, "empieza con un BOM; guárdalo como UTF-8 sin BOM")
	case !utf8.Valid(data):
		s.add("V38", rel, 0, "no es UTF-8 válido (¿Latin-1?); conviértelo a UTF-8")
		return
	case bytes.Contains(data, []byte("\r")):
		s.add("V38", rel, bytes.Count(data[:bytes.IndexByte(data, '\r')], []byte{'\n'})+1, "tiene fin de línea CRLF; usa LF")
	}
	latex := contract.IsLaTeXSource(name)
	for i, line := range strings.Split(string(data), "\n") {
		if reMarker.MatchString(line) {
			s.add("V39", rel, i+1, "usa {{TITULO}}/{{TITLE}}: el título es un parámetro con fromProjectName")
		}
		if !latex {
			continue
		}
		code := stripComment(line)
		if reOutsidePath.MatchString(code) {
			s.add("V35", rel, i+1, "carga un archivo fuera de la plantilla (../ o ruta absoluta)")
		}
		if reShellEscape.MatchString(code) {
			s.add("V39", rel, i+1, "necesita --shell-escape (ejecutar comandos del sistema), prohibido en v1")
		}
		for _, mm := range reProvide.FindAllStringSubmatch(code, -1) {
			s.macros[mm[1]] = true
		}
	}
}

// checkParameters applies V36 (ink-params before the defaults, in the entry)
// and V37 (every parameter macro has a \providecommand in src/).
func checkParameters(m *contract.Manifest, entryPath, entryRel string, macros map[string]bool,
	add func(code, field string, line int, format string, args ...any)) {
	if len(m.Parameters) == 0 {
		return
	}
	params := map[string]bool{}
	for _, p := range m.Parameters {
		params[p.Macro] = true
		if !macros[p.Macro] {
			add("V37", entryRel, 0, `falta \providecommand{\%s}: sin ink-params.tex la plantilla no compilaría`, p.Macro)
		}
	}
	data, err := os.ReadFile(entryPath)
	if err != nil {
		return // V09 already reports a missing entry
	}
	inkLine, firstDefault := 0, 0
	for i, line := range strings.Split(string(data), "\n") {
		code := stripComment(line)
		if inkLine == 0 && reInkParams.MatchString(code) {
			inkLine = i + 1
		}
		if firstDefault == 0 {
			for _, mm := range reProvide.FindAllStringSubmatch(code, -1) {
				if params[mm[1]] {
					firstDefault = i + 1
					break
				}
			}
		}
	}
	switch {
	case inkLine == 0:
		add("V36", entryRel, 0, `falta \InputIfFileExists{ink-params}{}{}: los parámetros nunca llegarían a la plantilla`)
	case firstDefault != 0 && inkLine > firstDefault:
		add("V36", entryRel, inkLine, `\InputIfFileExists{ink-params} va después del primer valor por defecto (línea %d): el parámetro nunca ganaría`, firstDefault)
	}
}

// stripComment removes a LaTeX comment: from the first % not escaped by an
// odd number of backslashes.
func stripComment(line string) string {
	for i := 0; i < len(line); i++ {
		if line[i] != '%' {
			continue
		}
		bs := 0
		for j := i - 1; j >= 0 && line[j] == '\\'; j-- {
			bs++
		}
		if bs%2 == 0 {
			return line[:i]
		}
	}
	return line
}

func whyForbidden(name string) string {
	switch {
	case contract.IsLatexmkrc(name):
		return "un .latexmkrc es código Perl que latexmk ejecuta"
	case name == contract.LicenseFileName || name == contract.MetaFileName:
		return "lo genera inkt bundle"
	case strings.HasPrefix(name, "."):
		return "archivo oculto"
	default:
		return "extensión fuera de la lista del contrato §9 B4, que distingue mayúsculas"
	}
}

func isFile(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.Mode().IsRegular()
}
