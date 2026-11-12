package contract

import (
	"regexp"
	"strings"
	"time"
)

// Patterns of the contract (§2–§4). Go's regexp is RE2, without lookahead:
// rules the JSON Schema expresses with (?!…) are written as code here.
var (
	reID          = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	reSemver      = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
	reCategory    = regexp.MustCompile(`^[a-z]+(-[a-z]+)*$`)
	reLanguage    = regexp.MustCompile(`^[a-z]{2}(-[A-Z]{2})?$`)
	reSpanish     = regexp.MustCompile(`^es(-[A-Z]{2})?$`)
	reLicenseID   = regexp.MustCompile(`^[A-Za-z0-9.+-]+$`)
	reGitHub      = regexp.MustCompile(`^[A-Za-z0-9-]+$`)
	reSegment     = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	reTeXPackage  = regexp.MustCompile(`^[a-z0-9-]+$`)
	reTeXFile     = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	reCreativeInk = regexp.MustCompile(`^>=(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
	reComponent   = regexp.MustCompile(`^[a-z0-9-]+(/[a-z0-9-]+)*$`)
	reKey         = regexp.MustCompile(`^[a-z][a-zA-Z0-9]{0,39}$`)
	reMacro       = regexp.MustCompile(`^[A-Za-z]{1,40}$`)
	reEnumValue   = regexp.MustCompile(`^[A-Za-z0-9-]{1,40}$`)
	reSHA256      = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// ValidID reports whether s is a valid template or pack id (§2).
func ValidID(s string) bool { return len(s) >= 2 && len(s) <= 40 && reID.MatchString(s) }

// ValidSemver reports whether s is MAJOR.MINOR.PATCH without prerelease (§2).
func ValidSemver(s string) bool { return reSemver.MatchString(s) }

// ValidSHA256 reports whether s is 64 lowercase hexadecimal digits.
func ValidSHA256(s string) bool { return reSHA256.MatchString(s) }

// ValidDate reports whether s is a real date in AAAA-MM-DD form.
func ValidDate(s string) bool {
	if len(s) != len("2006-01-02") {
		return false
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

// ValidRelPath reports whether p is a portable relative path (§2, B2): "/"
// separators, no leading "/", no "." or ".." segment, and every segment in
// [A-Za-z0-9._-]. It rejects "C:" and "\" on every OS, not only on Windows.
func ValidRelPath(p string) bool {
	if p == "" || strings.HasPrefix(p, "/") {
		return false
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == "." || seg == ".." || !reSegment.MatchString(seg) {
			return false
		}
	}
	return true
}

// ParseLicense splits an SPDX expression joined with " AND " (§3). It
// returns false for OR, WITH, parentheses or a malformed identifier.
func ParseLicense(expr string) ([]string, bool) {
	if strings.TrimSpace(expr) == "" {
		return nil, false
	}
	ids := strings.Split(expr, " AND ")
	for _, id := range ids {
		if !reLicenseID.MatchString(id) {
			return nil, false
		}
	}
	return ids, true
}

// reservedMacros cannot be used as parameter macros (§4.4). Macros starting
// with "ink"/"Ink" are reserved too (see IsReservedMacro).
var reservedMacros = map[string]bool{
	"borrador": true, "SoloCapitulo": true,
	"title": true, "author": true, "date": true, "thanks": true, "today": true, "maketitle": true,
	"documentclass": true, "usepackage": true, "begin": true, "end": true, "input": true,
	"include": true, "includeonly": true, "part": true, "chapter": true, "section": true,
	"subsection": true, "paragraph": true, "label": true, "ref": true, "pageref": true,
	"cite": true, "item": true, "caption": true, "footnote": true, "newcommand": true,
	"renewcommand": true, "providecommand": true, "def": true, "let": true, "relax": true,
	"par": true, "color": true, "textcolor": true, "LaTeX": true, "TeX": true,
}

// IsReservedMacro reports whether name is reserved for LaTeX, Creative-Ink or
// the shared components (§4.4).
func IsReservedMacro(name string) bool {
	return reservedMacros[name] || strings.HasPrefix(name, "ink") || strings.HasPrefix(name, "Ink")
}

func isRecommendedCategory(c string) bool {
	for _, r := range RecommendedCategories {
		if r == c {
			return true
		}
	}
	return false
}
