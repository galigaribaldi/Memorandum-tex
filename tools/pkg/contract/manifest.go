package contract

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"
)

// Origin says where a manifest comes from: the repository (no dist) or a
// published catalog (dist required).
type Origin int

// Origins.
const (
	OriginRepository Origin = iota
	OriginPublished
)

// Options of ValidateManifest.
type Options struct {
	Origin Origin
	// RepoPolicy adds the rules of this repository on top of the contract:
	// Spanish, pdflatex and xelatex with pdflatex by default, MIT-0, and a
	// warning for a category outside RecommendedCategories. Catalogs of third
	// parties are validated without it.
	RepoPolicy bool
	// Context enables the checks that need the template folder. Nil skips them.
	Context *Context
}

// Context locates a template on disk. Empty fields skip their checks.
type Context struct {
	TemplateDir string // folder with template.json and src/: entry (V09) and preview (V17)
	FolderName  string // expected id (V04)
	ParentName  string // expected category (V04)
	SharedDir   string // shared/: components (V18)
	LicensesDir string // LICENSES/: license texts (V08)
}

// knownManifestFields are the top-level keys of §3 (anything else is V20).
var knownManifestFields = []string{
	"$schema", "schemaVersion", "id", "version", "name", "description", "category", "tags", "language",
	"license", "authors", "entry", "build", "requires", "components", "parameters", "modes", "preview", "dist",
}

var (
	knownBuildFields     = []string{"engines", "defaultEngine", "bibliography", "shellEscape"}
	knownRequiresFields  = []string{"creativeInk", "texLive", "packages", "files"}
	knownAuthorFields    = []string{"name", "github", "email"}
	knownParameterFields = []string{"key", "macro", "type", "label", "description", "group", "required", "default",
		"maxLength", "maxItems", "fromProjectName", "values"}
	knownDistFields = []string{"url", "sha256", "size", "publishedAt"}
)

// requiredManifestFields and the rule that reports each one missing.
var requiredManifestFields = []struct{ field, code string }{
	{"id", "V03"}, {"version", "V05"}, {"name", "V06"}, {"description", "V06"}, {"category", "V06"},
	{"language", "V07"}, {"license", "V08"}, {"authors", "V08"}, {"entry", "V09"}, {"build", "V10"}, {"modes", "V16"},
}

// fieldCodes maps a top-level key to the rule that owns it, for type errors.
var fieldCodes = map[string]string{
	"schemaVersion": "V02", "id": "V03", "version": "V05", "name": "V06", "description": "V06",
	"category": "V06", "tags": "V06", "language": "V07", "license": "V08", "authors": "V08", "entry": "V09",
	"build": "V10", "requires": "V11", "parameters": "V13", "modes": "V16", "preview": "V17",
	"components": "V18", "dist": "V19",
}

// ValidateManifest parses and validates template.json (§3–§5, rules V01–V20).
// The manifest is nil when it cannot be read at all (V01) or uses another
// schemaVersion (V02); otherwise it is returned even if it has errors.
func ValidateManifest(data []byte, opt Options) (*Manifest, Report) {
	r := newReporter(data)
	raw, ok := checkJSON(r, data, MaxManifestBytes)
	if !ok {
		return nil, r.report
	}
	if !checkSchemaVersion(r, raw) {
		return nil, r.report
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		var te *json.UnmarshalTypeError
		if errors.As(err, &te) {
			code := fieldCodes[strings.SplitN(te.Field, ".", 2)[0]]
			if strings.HasPrefix(te.Field, "parameters.") {
				code = parameterFieldCode(te.Field)
			}
			if code == "" {
				code = "V01"
			}
			r.errorf(code, te.Field, "tipo incorrecto: se esperaba %s", te.Type)
		} else {
			r.errorf("V01", "", "no se pudo leer el manifiesto: %v", err)
			return nil, r.report
		}
	}

	for _, req := range requiredManifestFields {
		if _, ok := raw[req.field]; !ok {
			r.errorf(req.code, req.field, "campo obligatorio ausente")
		}
	}
	checkUnknownFields(r, raw)

	v := manifestValidator{r: r, m: &m, raw: raw, opt: opt}
	v.identity()
	v.license()
	v.entry()
	v.build()
	v.requires()
	v.parameters()
	v.modes()
	v.preview()
	v.components()
	v.dist()
	return &m, r.report
}

// checkJSON applies V01: size, BOM, UTF-8 and a JSON object.
func checkJSON(r *reporter, data []byte, limit int) (map[string]json.RawMessage, bool) {
	switch {
	case len(data) > limit:
		r.errorf("V01", "", "el archivo ocupa %d bytes; el máximo es %d", len(data), limit)
		return nil, false
	case bytes.HasPrefix(data, []byte("\xef\xbb\xbf")):
		r.errorf("V01", "", "el archivo empieza con un BOM; guárdalo como UTF-8 sin BOM")
		return nil, false
	case !utf8.Valid(data):
		r.errorf("V01", "", "el archivo no es UTF-8 válido")
		return nil, false
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		var se *json.SyntaxError
		if errors.As(err, &se) {
			r.minimum = bytes.Count(data[:min(int(se.Offset), len(data))], []byte{'\n'}) + 1
			r.errorf("V01", "", "JSON inválido: %v", err)
		} else {
			r.errorf("V01", "", "el manifiesto debe ser un objeto JSON")
		}
		return nil, false
	}
	return raw, true
}

// checkSchemaVersion applies V02. A different version stops the validation:
// the other rules belong to another contract.
func checkSchemaVersion(r *reporter, raw map[string]json.RawMessage) bool {
	sv, ok := raw["schemaVersion"]
	if !ok {
		r.errorf("V02", "schemaVersion", "campo obligatorio ausente")
		return false
	}
	var n int
	if err := json.Unmarshal(sv, &n); err != nil {
		r.errorf("V02", "schemaVersion", "debe ser un entero")
		return false
	}
	if n != SchemaVersion {
		r.errorf("V02", "schemaVersion", "es %d; esta versión de la herramienta implementa el contrato %d", n, SchemaVersion)
		return false
	}
	return true
}

func parameterFieldCode(field string) string {
	switch strings.TrimPrefix(field, "parameters.") {
	case "key", "macro":
		return "V12"
	case "default":
		return "V14"
	case "fromProjectName":
		return "V15"
	}
	return "V13"
}

// checkUnknownFields applies V20 at every level of the manifest.
func checkUnknownFields(r *reporter, raw map[string]json.RawMessage) {
	warnUnknown(r, "", raw, knownManifestFields)
	nested := func(field string, known []string) {
		var obj map[string]json.RawMessage
		if json.Unmarshal(raw[field], &obj) == nil {
			warnUnknown(r, field, obj, known)
		}
	}
	nested("build", knownBuildFields)
	nested("requires", knownRequiresFields)
	nested("dist", knownDistFields)
	list := func(field string, known []string) {
		var items []map[string]json.RawMessage
		if json.Unmarshal(raw[field], &items) == nil {
			for i, obj := range items {
				warnUnknown(r, fmt.Sprintf("%s[%d]", field, i), obj, known)
			}
		}
	}
	list("authors", knownAuthorFields)
	list("parameters", knownParameterFields)
}

func warnUnknown(r *reporter, prefix string, obj map[string]json.RawMessage, known []string) {
	keys := make([]string, 0, len(obj))
	for k := range obj {
		if !slices.Contains(known, k) {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)
	for _, k := range keys {
		field := k
		if prefix != "" {
			field = prefix + "." + k
		}
		r.warnf("V20", field, "campo desconocido (¿una errata?); los clientes lo ignoran")
	}
}

type manifestValidator struct {
	r   *reporter
	m   *Manifest
	raw map[string]json.RawMessage
	opt Options
}

func (v *manifestValidator) has(field string) bool { _, ok := v.raw[field]; return ok }

func (v *manifestValidator) ctx() *Context {
	if v.opt.Context == nil {
		return &Context{}
	}
	return v.opt.Context
}

// identity applies V03–V07.
func (v *manifestValidator) identity() {
	m, r := v.m, v.r
	if v.has("id") && !ValidID(m.ID) {
		r.errorf("V03", "id", "%q no es un id válido: kebab-case ASCII de 2 a 40 caracteres", m.ID)
	}
	// V04 only compares a valid id: an invalid one is already V03.
	if c := v.ctx(); c.FolderName != "" && v.has("id") && ValidID(m.ID) && m.ID != c.FolderName {
		r.errorf("V04", "id", "es %q pero la carpeta se llama %q", m.ID, c.FolderName)
	}
	if v.has("version") && !ValidSemver(m.Version) {
		r.errorf("V05", "version", "%q no es semver MAJOR.MINOR.PATCH (sin prerelease)", m.Version)
	}
	if v.has("name") {
		if n := utf8.RuneCountInString(m.Name); n < 1 || n > 80 {
			r.errorf("V06", "name", "debe tener de 1 a 80 caracteres (tiene %d)", n)
		}
	}
	if v.has("description") {
		if n := utf8.RuneCountInString(m.Description); n < 1 || n > 300 {
			r.errorf("V06", "description", "debe tener de 1 a 300 caracteres (tiene %d)", n)
		} else if strings.ContainsAny(m.Description, "\r\n") {
			r.errorf("V06", "description", "debe ser una sola línea")
		}
	}
	if v.has("category") {
		switch {
		case !reCategory.MatchString(m.Category):
			r.errorf("V06", "category", "%q no es válida: minúsculas ASCII separadas por guiones", m.Category)
		case v.opt.RepoPolicy && !isRecommendedCategory(m.Category):
			r.warnf("V06", "category", "%q no es una categoría recomendada (%s)", m.Category, strings.Join(RecommendedCategories, ", "))
		}
	}
	if c := v.ctx(); c.ParentName != "" && v.has("category") && reCategory.MatchString(m.Category) && m.Category != c.ParentName {
		r.errorf("V04", "category", "es %q pero la carpeta padre se llama %q", m.Category, c.ParentName)
	}
	if len(m.Tags) > 10 {
		v.r.errorf("V06", "tags", "hay %d etiquetas; el máximo es 10", len(m.Tags))
	}
	seen := map[string]bool{}
	for i, t := range m.Tags {
		f := fmt.Sprintf("tags[%d]", i)
		switch {
		case !reID.MatchString(t):
			r.errorf("V06", f, "%q no es válida: kebab-case ASCII", t)
		case seen[t]:
			r.errorf("V06", f, "%q está repetida", t)
		}
		seen[t] = true
	}
	if v.has("language") {
		switch {
		case !reLanguage.MatchString(m.Language):
			r.errorf("V07", "language", "%q no es un código de idioma (es, es-MX)", m.Language)
		case v.opt.RepoPolicy && !reSpanish.MatchString(m.Language):
			r.errorf("V07", "language", "las plantillas de este repositorio están en español (es o es-XX)")
		}
	}
}

// license applies V08 (license and authors).
func (v *manifestValidator) license() {
	m, r := v.m, v.r
	if v.has("license") {
		ids, ok := ParseLicense(m.License)
		switch {
		case !ok:
			r.errorf("V08", "license", "%q no es una expresión SPDX válida (identificadores unidos con \" AND \")", m.License)
		default:
			if v.opt.RepoPolicy && !slices.Contains(ids, PolicyLicense) {
				r.errorf("V08", "license", "las plantillas de este repositorio incluyen %s (NOTICE.md)", PolicyLicense)
			}
			if dir := v.ctx().LicensesDir; dir != "" {
				for _, id := range ids {
					if !isFile(filepath.Join(dir, id+".txt")) {
						r.errorf("V08", "license", "falta el texto de %s en LICENSES/%s.txt", id, id)
					}
				}
			}
		}
	}
	if v.has("authors") && len(m.Authors) == 0 {
		r.errorf("V08", "authors", "debe haber al menos un autor")
	}
	for i, a := range m.Authors {
		f := fmt.Sprintf("authors[%d]", i)
		if strings.TrimSpace(a.Name) == "" {
			r.errorf("V08", f+".name", "el nombre es obligatorio")
		}
		if a.GitHub != "" && !reGitHub.MatchString(a.GitHub) {
			r.errorf("V08", f+".github", "%q no es un usuario de GitHub válido", a.GitHub)
		}
	}
}

// entry applies V09.
func (v *manifestValidator) entry() {
	if !v.has("entry") {
		return
	}
	e := v.m.Entry
	if !ValidRelPath(e) || !strings.HasSuffix(e, ".tex") {
		v.r.errorf("V09", "entry", "%q debe ser una ruta relativa a src/ terminada en .tex, sin '..'", e)
		return
	}
	if dir := v.ctx().TemplateDir; dir != "" && !isFile(filepath.Join(dir, "src", filepath.FromSlash(e))) {
		v.r.errorf("V09", "entry", "no existe src/%s", e)
	}
}

// build applies V10.
func (v *manifestValidator) build() {
	if !v.has("build") {
		return
	}
	b, r := v.m.Build, v.r
	known := []string{EnginePDFLaTeX, EngineXeLaTeX, EngineLuaLaTeX}
	if len(b.Engines) == 0 {
		r.errorf("V10", "build.engines", "debe listar al menos un motor")
	}
	seen := map[string]bool{}
	for i, e := range b.Engines {
		f := fmt.Sprintf("build.engines[%d]", i)
		switch {
		case !slices.Contains(known, e):
			r.errorf("V10", f, "%q no es un motor (%s)", e, strings.Join(known, ", "))
		case seen[e]:
			r.errorf("V10", f, "%q está repetido", e)
		}
		seen[e] = true
	}
	if !slices.Contains(b.Engines, b.DefaultEngine) {
		r.errorf("V10", "build.defaultEngine", "%q debe ser uno de build.engines", b.DefaultEngine)
	}
	if !slices.Contains([]string{BibliographyNone, BibliographyBibTeX, BibliographyBiber}, b.Bibliography) {
		r.errorf("V10", "build.bibliography", "%q no es válido (none, bibtex, biber)", b.Bibliography)
	}
	if v.opt.RepoPolicy {
		if !slices.Contains(b.Engines, EnginePDFLaTeX) || !slices.Contains(b.Engines, EngineXeLaTeX) {
			r.errorf("V10", "build.engines", "las plantillas de este repositorio compilan con pdflatex y xelatex")
		}
		if b.DefaultEngine != EnginePDFLaTeX && slices.Contains(b.Engines, b.DefaultEngine) {
			r.errorf("V10", "build.defaultEngine", "el motor por defecto de este repositorio es pdflatex")
		}
		// Third-party catalogs are not rejected here: compat blocks them.
		if b.ShellEscape {
			r.errorf("V10", "build.shellEscape", "prohibido en v1: --shell-escape permite ejecutar comandos del sistema")
		}
	}
}

// requires applies V11.
func (v *manifestValidator) requires() {
	q, r := v.m.Requires, v.r
	if q == nil {
		return
	}
	if q.CreativeInk != "" {
		if _, ok := MinimumCreativeInk(q.CreativeInk); !ok {
			r.errorf("V11", "requires.creativeInk", "%q debe tener la forma >=X.Y.Z", q.CreativeInk)
		}
	}
	if q.TeXLive != 0 && (q.TeXLive < 2020 || q.TeXLive > 2100) {
		r.errorf("V11", "requires.texLive", "%d no es un año de TeX Live entre 2020 y 2100", q.TeXLive)
	}
	checkNames := func(field string, list []string, ok func(string) bool, what string) {
		seen := map[string]bool{}
		for i, s := range list {
			f := fmt.Sprintf("%s[%d]", field, i)
			switch {
			case !ok(s):
				r.errorf("V11", f, "%q no es un nombre de %s válido", s, what)
			case seen[s]:
				r.errorf("V11", f, "%q está repetido", s)
			}
			seen[s] = true
		}
	}
	checkNames("requires.packages", q.Packages, reTeXPackage.MatchString, "paquete de tlmgr")
	checkNames("requires.files", q.Files, reTeXFile.MatchString, "archivo para kpsewhich")
}

// modes applies V16.
func (v *manifestValidator) modes() {
	if !v.has("modes") {
		return
	}
	known := []string{ModeCompleto, ModeRapido, ModeBorrador, ModeCapitulo}
	seen := map[string]bool{}
	for i, mode := range v.m.Modes {
		f := fmt.Sprintf("modes[%d]", i)
		switch {
		case !slices.Contains(known, mode):
			v.r.errorf("V16", f, "%q no es un modo (%s)", mode, strings.Join(known, ", "))
		case seen[mode]:
			v.r.errorf("V16", f, "%q está repetido", mode)
		}
		seen[mode] = true
	}
	if !slices.Contains(v.m.Modes, ModeCompleto) {
		v.r.errorf("V16", "modes", "debe incluir %q", ModeCompleto)
	}
}

// preview applies V17.
func (v *manifestValidator) preview() {
	p := v.m.Preview
	if !v.has("preview") {
		return
	}
	if !ValidRelPath(p) || !strings.HasSuffix(p, ".png") {
		v.r.errorf("V17", "preview", "%q debe ser un PNG con ruta relativa a la carpeta de la plantilla", p)
		return
	}
	dir := v.ctx().TemplateDir
	if dir == "" {
		return
	}
	info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(p)))
	switch {
	case err != nil || !info.Mode().IsRegular():
		v.r.errorf("V17", "preview", "no existe %s", p)
	case info.Size() > MaxPreviewBytes:
		v.r.errorf("V17", "preview", "ocupa %d KB; el máximo es %d KB", info.Size()>>10, MaxPreviewBytes>>10)
	}
}

// components applies V18.
func (v *manifestValidator) components() {
	shared := v.ctx().SharedDir
	seen := map[string]bool{}
	for i, c := range v.m.Components {
		f := fmt.Sprintf("components[%d]", i)
		switch {
		case !reComponent.MatchString(c):
			v.r.errorf("V18", f, "%q no es un componente válido (p. ej. comandos/academico)", c)
		case seen[c]:
			v.r.errorf("V18", f, "%q está repetido", c)
		case shared != "" && !isDir(filepath.Join(shared, filepath.FromSlash(c))):
			v.r.errorf("V18", f, "no existe shared/%s", c)
		}
		seen[c] = true
	}
}

// dist applies V19: forbidden in the repository, required when published.
func (v *manifestValidator) dist() {
	d := v.m.Dist
	if v.opt.Origin == OriginRepository {
		if v.has("dist") {
			v.r.errorf("V19", "dist", "no va en el repositorio: lo añade la publicación")
		}
		return
	}
	if d == nil {
		v.r.errorf("V19", "dist", "un manifiesto publicado debe incluir dist")
		return
	}
	if !strings.HasPrefix(d.URL, "https://") {
		v.r.errorf("V19", "dist.url", "%q debe ser una URL https", d.URL)
	}
	if !ValidSHA256(d.SHA256) {
		v.r.errorf("V19", "dist.sha256", "debe tener 64 dígitos hexadecimales en minúscula")
	}
	if d.Size < 1 || d.Size > MaxBundleBytes {
		v.r.errorf("V19", "dist.size", "%d no está entre 1 y %d bytes", d.Size, MaxBundleBytes)
	}
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
