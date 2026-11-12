package contract

import "encoding/json"

// SchemaVersion is the version of the contract implemented by this package.
const SchemaVersion = 1

// Engines (build.engines, build.defaultEngine).
const (
	EnginePDFLaTeX = "pdflatex"
	EngineXeLaTeX  = "xelatex"
	EngineLuaLaTeX = "lualatex"
)

// Bibliography backends (build.bibliography).
const (
	BibliographyNone   = "none"
	BibliographyBibTeX = "bibtex"
	BibliographyBiber  = "biber"
)

// Modes (§5). The macros \borrador and \SoloCapitulo are fixed by Creative-Ink.
const (
	ModeCompleto = "completo"
	ModeRapido   = "rapido"
	ModeBorrador = "borrador"
	ModeCapitulo = "capitulo"
)

// ParamType is the type of a parameter (§4.2).
type ParamType string

// Parameter types.
const (
	TypeString  ParamType = "string"
	TypeBoolean ParamType = "boolean"
	TypeEnum    ParamType = "enum"
	TypeDate    ParamType = "date"
	TypeList    ParamType = "list"
)

// DefaultToday is the special default of a date parameter: the local date of
// whoever creates the project.
const DefaultToday = "today"

// PolicyLicense must appear in the license of every template of this
// repository (NOTICE.md).
const PolicyLicense = "MIT-0"

// Limits of the contract.
const (
	MaxManifestBytes       = 256 << 10 // V01
	MaxIndexBytes          = 1 << 20   // §7
	MaxBundleBytes         = 200 << 20 // B5, uncompressed
	MaxBundleFiles         = 2000      // B5
	MaxCompressionRatio    = 100       // B5, per file
	MaxPreviewBytes        = 200 << 10 // V17
	MaxParameters          = 40
	DefaultStringMaxLength = 300
	DefaultListMaxItems    = 20
)

// RecommendedCategories are the categories of this repository; another one
// is a warning (V06), not an error, so adding a category needs no contract change.
var RecommendedCategories = []string{"academico", "presentacion", "difusion", "correspondencia", "congresos"}

// Manifest is template.json (§3). In the repository Dist is absent; in a
// published manifest it is required (§6).
type Manifest struct {
	Schema        string      `json:"$schema,omitempty"`
	SchemaVersion int         `json:"schemaVersion"`
	ID            string      `json:"id"`
	Version       string      `json:"version"`
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	Category      string      `json:"category"`
	Tags          []string    `json:"tags,omitempty"`
	Language      string      `json:"language"`
	License       string      `json:"license"`
	Authors       []Author    `json:"authors"`
	Entry         string      `json:"entry"`
	Build         Build       `json:"build"`
	Requires      *Requires   `json:"requires,omitempty"`
	Components    []string    `json:"components,omitempty"`
	Parameters    []Parameter `json:"parameters,omitempty"`
	Modes         []string    `json:"modes"`
	Preview       string      `json:"preview,omitempty"`
	Dist          *Dist       `json:"dist,omitempty"`
}

// Author of a template.
type Author struct {
	Name   string `json:"name"`
	GitHub string `json:"github,omitempty"`
	Email  string `json:"email,omitempty"`
}

// Build is how the template compiles (§3.1).
type Build struct {
	Engines       []string `json:"engines"`
	DefaultEngine string   `json:"defaultEngine"`
	Bibliography  string   `json:"bibliography"`
	ShellEscape   bool     `json:"shellEscape,omitempty"`
}

// Requires lists what the template needs (§3.2).
type Requires struct {
	CreativeInk string   `json:"creativeInk,omitempty"`
	TeXLive     int      `json:"texLive,omitempty"`
	Packages    []string `json:"packages,omitempty"`
	Files       []string `json:"files,omitempty"`
}

// Parameter is one customizable value of a template (§4). Optional numeric
// and boolean fields are pointers so "absent" and "zero" differ.
type Parameter struct {
	Key             string          `json:"key"`
	Macro           string          `json:"macro"`
	Type            ParamType       `json:"type"`
	Label           string          `json:"label"`
	Description     string          `json:"description,omitempty"`
	Group           string          `json:"group,omitempty"`
	Required        bool            `json:"required,omitempty"`
	Default         json.RawMessage `json:"default,omitempty"`
	MaxLength       *int            `json:"maxLength,omitempty"`
	MaxItems        *int            `json:"maxItems,omitempty"`
	FromProjectName *bool           `json:"fromProjectName,omitempty"`
	Values          []string        `json:"values,omitempty"`
}

// EffectiveMaxLength is maxLength or its default (strings and list items).
func (p Parameter) EffectiveMaxLength() int {
	if p.MaxLength != nil {
		return *p.MaxLength
	}
	return DefaultStringMaxLength
}

// EffectiveMaxItems is maxItems or its default (lists).
func (p Parameter) EffectiveMaxItems() int {
	if p.MaxItems != nil {
		return *p.MaxItems
	}
	return DefaultListMaxItems
}

// UsesProjectName reports fromProjectName: true.
func (p Parameter) UsesProjectName() bool { return p.FromProjectName != nil && *p.FromProjectName }

// HasDefault reports whether the parameter declares a default (null counts as absent).
func (p Parameter) HasDefault() bool {
	return len(p.Default) > 0 && string(p.Default) != "null"
}

// Dist is where a published version is downloaded from (§6).
type Dist struct {
	URL         string `json:"url"`
	SHA256      string `json:"sha256"`
	Size        int64  `json:"size"`
	PublishedAt string `json:"publishedAt,omitempty"`
}

// Index is catalog/v1/index.json (§7).
type Index struct {
	SchemaVersion int             `json:"schemaVersion"`
	GeneratedAt   string          `json:"generatedAt"`
	Source        Source          `json:"source"`
	Templates     []IndexTemplate `json:"templates"`
	Packs         []IndexPack     `json:"packs"`
}

// Source identifies a catalog.
type Source struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Homepage string `json:"homepage,omitempty"`
}

// IndexTemplate summarizes the latest version of a template.
type IndexTemplate struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	Tags          []string  `json:"tags,omitempty"`
	Language      string    `json:"language"`
	License       string    `json:"license"`
	Latest        string    `json:"latest"`
	Engines       []string  `json:"engines"`
	DefaultEngine string    `json:"defaultEngine"`
	Bibliography  string    `json:"bibliography"`
	ShellEscape   bool      `json:"shellEscape"`
	Requires      *Requires `json:"requires,omitempty"`
	Manifest      string    `json:"manifest"`
	Versions      string    `json:"versions"`
	Preview       string    `json:"preview,omitempty"`
	Dist          Dist      `json:"dist"`
}

// IndexPack points to a pack.
type IndexPack struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Manifest    string `json:"manifest"`
}

// Versions is catalog/v1/templates/<id>/versions.json (§6), newest first.
type Versions struct {
	SchemaVersion int            `json:"schemaVersion"`
	ID            string         `json:"id"`
	Versions      []VersionEntry `json:"versions"`
}

// VersionEntry is one published version.
type VersionEntry struct {
	Version     string `json:"version"`
	Manifest    string `json:"manifest"`
	PublishedAt string `json:"publishedAt,omitempty"`
}

// Pack is packs/<id>.json (§8).
type Pack struct {
	Schema        string         `json:"$schema,omitempty"`
	SchemaVersion int            `json:"schemaVersion"`
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Templates     []PackTemplate `json:"templates"`
}

// PackTemplate is one entry of a pack; an empty Version means latest.
type PackTemplate struct {
	ID      string `json:"id"`
	Version string `json:"version,omitempty"`
}

// ReservedPackID is generated by inkt index with every template (§8).
const ReservedPackID = "todo"
