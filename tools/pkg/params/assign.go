package params

import (
	"fmt"
	"strings"
)

// ParseAssignments turns command-line words such as `portada=false` or
// `titulo=Mi tesis` into Input.Values. The key ends at the first "=", so a
// value may contain "=". An empty key, a word without "=" or a key given twice
// is an error: silently keeping the last one would hide a typo.
func ParseAssignments(words []string) (map[string]string, error) {
	values := make(map[string]string, len(words))
	for _, w := range words {
		key, value, ok := strings.Cut(w, "=")
		key = strings.TrimSpace(key)
		switch {
		case !ok:
			return nil, fmt.Errorf("%q no tiene la forma clave=valor", w)
		case key == "":
			return nil, fmt.Errorf("%q no tiene clave antes del =", w)
		}
		if _, dup := values[key]; dup {
			return nil, fmt.Errorf("el parámetro %q aparece dos veces", key)
		}
		values[key] = value
	}
	return values, nil
}
