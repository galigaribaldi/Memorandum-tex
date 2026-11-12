// Package fixtures embeds the test data of the template contract v1
// (data/README.md, docs/contrato-plantillas.md). It lives inside the tools
// module so that it travels with it: Creative-Ink gets the same fixtures when
// it imports the module with go get.
package fixtures

import (
	"embed"
	"io/fs"
	"path/filepath"
	"runtime"
)

// data holds every fixture. The "all:" prefix keeps files whose names start
// with "." or "_", should a fixture ever need one.
//
//go:embed all:data
var data embed.FS

// FS returns the fixtures rooted at data/: "templates/prueba/ok-minima/template.json",
// "params/escape-cases.json", and so on.
func FS() fs.FS {
	sub, err := fs.Sub(data, "data")
	if err != nil {
		panic(err) // the embed pattern guarantees data/ exists
	}
	return sub
}

// Dir returns the absolute path of the data/ folder on disk, for checks that
// need real files (an entry that exists, a preview's size). It is only valid
// when the module's sources are on disk (this repository, a replace directive
// or the module cache); ok is false otherwise.
func Dir() (dir string, ok bool) {
	_, file, _, ok := runtime.Caller(0)
	if !ok || !filepath.IsAbs(file) { // -trimpath leaves a module-relative path
		return "", false
	}
	return filepath.Join(filepath.Dir(file), "data"), true
}
