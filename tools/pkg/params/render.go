package params

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/galigaribaldi/Memorandum-tex/tools/pkg/contract"
)

// FileName is the file that templates load with \InputIfFileExists{ink-params}.
const FileName = "ink-params.tex"

// texEscaper matches core.EscapeTeX of Creative-Ink: one pass, so the braces
// of \textbackslash{} are not escaped again (§10, params/escape-cases.json).
var texEscaper = strings.NewReplacer(
	`\`, `\textbackslash{}`, "{", `\{`, "}", `\}`, "$", `\$`, "&", `\&`,
	"#", `\#`, "%", `\%`, "_", `\_`, "~", `\textasciitilde{}`, "^", `\textasciicircum{}`,
)

// EscapeTeX makes s safe as the body of a LaTeX macro.
func EscapeTeX(s string) string { return texEscaper.Replace(s) }

// RenderInkParams writes ink-params.tex (§10): two header lines and one
// \def\<macro>{<value>} per resolved parameter, in manifest order, with LF.
// The output is byte-for-byte stable: the tests compare it with golden files.
func RenderInkParams(r *Resolved) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%% %s — parámetros de %s %s.\n", FileName, r.ID, r.Version)
	b.WriteString("% Generado a partir de template.json; puedes editar los valores.\n")
	for _, it := range r.Items {
		fmt.Fprintf(&b, "\\def\\%s{%s}\n", it.Param.Macro, texValue(it))
	}
	return b.Bytes()
}

func texValue(it Item) string {
	switch it.Param.Type {
	case contract.TypeBoolean:
		if it.Bool {
			return "true"
		}
		return "false"
	case contract.TypeList:
		var b strings.Builder
		for _, s := range it.List {
			b.WriteString(`\inkItem{` + EscapeTeX(s) + "}")
		}
		return b.String()
	case contract.TypeString:
		return EscapeTeX(it.String)
	default: // enum and date: restricted alphabets, written as they are
		return it.String
	}
}
