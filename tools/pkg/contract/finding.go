package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Level of a Finding: an error makes the manifest invalid, a warning does not.
type Level string

// Levels.
const (
	LevelError   Level = "error"
	LevelWarning Level = "warning"
)

// Finding is one broken rule. Code is the stable rule code of §12 (V01…V52).
// Field is a path into the JSON ("parameters[2].macro", "" for the whole
// file) and Line its 1-based line in the source (0 when unknown).
type Finding struct {
	Code    string `json:"code"`
	Level   Level  `json:"level"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
}

// String formats the finding as "V12 error: parameters[0].macro: …".
func (f Finding) String() string {
	if f.Field == "" {
		return fmt.Sprintf("%s %s: %s", f.Code, f.Level, f.Message)
	}
	return fmt.Sprintf("%s %s: %s: %s", f.Code, f.Level, f.Field, f.Message)
}

// Report is the result of a validation, in the order the rules ran.
type Report struct {
	Findings []Finding `json:"findings"`
}

// HasErrors reports whether any finding is an error.
func (r Report) HasErrors() bool {
	for _, f := range r.Findings {
		if f.Level == LevelError {
			return true
		}
	}
	return false
}

// Errors returns the findings of level error.
func (r Report) Errors() []Finding { return r.filter(LevelError) }

// Warnings returns the findings of level warning.
func (r Report) Warnings() []Finding { return r.filter(LevelWarning) }

func (r Report) filter(l Level) []Finding {
	var out []Finding
	for _, f := range r.Findings {
		if f.Level == l {
			out = append(out, f)
		}
	}
	return out
}

// reporter collects findings and resolves their lines.
type reporter struct {
	report  Report
	lines   map[string]int
	minimum int // line used when a field cannot be located
}

func newReporter(data []byte) *reporter {
	return &reporter{lines: locate(data), minimum: 1}
}

func (r *reporter) add(code string, level Level, field, format string, args ...any) {
	r.report.Findings = append(r.report.Findings, Finding{
		Code: code, Level: level, Field: field, Message: fmt.Sprintf(format, args...), Line: r.line(field),
	})
}

func (r *reporter) errorf(code, field, format string, args ...any) {
	r.add(code, LevelError, field, format, args...)
}

func (r *reporter) warnf(code, field, format string, args ...any) {
	r.add(code, LevelWarning, field, format, args...)
}

// line finds field or, if it is absent (a missing key), its closest parent.
func (r *reporter) line(field string) int {
	for f := field; f != ""; f = parentField(f) {
		if n, ok := r.lines[f]; ok {
			return n
		}
	}
	return r.minimum
}

func parentField(f string) string {
	if strings.HasSuffix(f, "]") {
		if i := strings.LastIndexByte(f, '['); i >= 0 {
			return f[:i]
		}
	}
	if i := strings.LastIndexByte(f, '.'); i >= 0 {
		return f[:i]
	}
	return ""
}

// locate maps every key and array element of a JSON document to the line
// where it starts ("parameters[2].macro" → 31). Errors leave the map partial:
// syntax errors are reported separately (V01).
func locate(data []byte) map[string]int {
	lines := map[string]int{}
	dec := json.NewDecoder(bytes.NewReader(data))
	lineAt := func(off int64) int {
		// InputOffset points right after the previous token: skip the
		// separators so the line is the one where the value starts.
		for off < int64(len(data)) && strings.IndexByte(" \t\r\n,:", data[off]) >= 0 {
			off++
		}
		return bytes.Count(data[:off], []byte{'\n'}) + 1
	}
	var walk func(path string) bool
	walk = func(path string) bool {
		tok, err := dec.Token()
		if err != nil {
			return false
		}
		delim, ok := tok.(json.Delim)
		if !ok {
			return true
		}
		switch delim {
		case '{':
			for dec.More() {
				off := dec.InputOffset()
				k, err := dec.Token()
				if err != nil {
					return false
				}
				key, _ := k.(string)
				child := key
				if path != "" {
					child = path + "." + key
				}
				lines[child] = lineAt(off)
				if !walk(child) {
					return false
				}
			}
		case '[':
			for i := 0; dec.More(); i++ {
				child := path + "[" + strconv.Itoa(i) + "]"
				lines[child] = lineAt(dec.InputOffset())
				if !walk(child) {
					return false
				}
			}
		}
		_, err = dec.Token() // closing delimiter
		return err == nil
	}
	walk("")
	return lines
}
