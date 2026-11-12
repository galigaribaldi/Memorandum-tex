package contract_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/galigaribaldi/Memorandum-tex/tools/internal/fixturetest"
	"github.com/galigaribaldi/Memorandum-tex/tools/pkg/contract"
)

// TestRepositoryPacks validates packs/*.json as inkt validate does today:
// templates that do not exist yet are warnings (V52), not errors.
func TestRepositoryPacks(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(fixturetest.RepoRoot(t), "packs", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no hay paquetes en packs/: %v", err)
	}
	for _, path := range files {
		name := strings.TrimSuffix(filepath.Base(path), ".json")
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			p, rep := contract.ValidatePack(data, contract.PackOptions{FileName: name, Known: func(string) bool { return false }})
			if p == nil || rep.HasErrors() {
				t.Fatalf("errores: %v", rep.Findings)
			}
			if len(rep.Warnings()) != len(p.Templates) {
				t.Errorf("se esperaba un V52 por plantilla aún inexistente: %v", rep.Findings)
			}
		})
	}
}

func TestPackRules(t *testing.T) {
	data := readFixture(t, "packs/prueba.json")
	known := func(id string) bool { return id == "ok-minima" || id == "con-paquetes" }

	if _, rep := contract.ValidatePack(data, contract.PackOptions{FileName: "prueba", Known: known}); len(rep.Findings) != 0 {
		t.Errorf("packs/prueba.json debe ser válido: %v", rep.Findings)
	}
	if _, rep := contract.ValidatePack(data, contract.PackOptions{FileName: "otro"}); !keysOf(rep)[findingKey{"V50", contract.LevelError}] {
		t.Errorf("un nombre de archivo distinto del id debe dar V50: %v", rep.Findings)
	}
	none := func(string) bool { return false }
	if _, rep := contract.ValidatePack(data, contract.PackOptions{Known: none, MissingIsError: true}); len(rep.Errors()) != 2 {
		t.Errorf("con MissingIsError, cada plantilla inexistente es un error V52: %v", rep.Findings)
	}

	todo := []byte(`{"schemaVersion":1,"id":"todo","name":"Todas","description":"Todas.","templates":[{"id":"ok-minima"}]}`)
	if _, rep := contract.ValidatePack(todo, contract.PackOptions{}); !keysOf(rep)[findingKey{"V50", contract.LevelError}] {
		t.Errorf("'todo' está reservado: %v", rep.Findings)
	}
	if _, rep := contract.ValidatePack(todo, contract.PackOptions{AllowReserved: true}); rep.HasErrors() {
		t.Errorf("inkt index puede generar 'todo': %v", rep.Findings)
	}

	bad := []byte(`{"schemaVersion":1,"id":"x-y","name":"","description":"a\nb","templates":[{"id":"A"},{"id":"ok","version":"1.0"},{"id":"ok"}]}`)
	_, rep := contract.ValidatePack(bad, contract.PackOptions{})
	if n := len(rep.Errors()); n != 5 { // name, description, id "A", version "1.0", "ok" repetida
		t.Errorf("se esperaban 5 errores V51, hubo %d: %v", n, rep.Findings)
	}
}
