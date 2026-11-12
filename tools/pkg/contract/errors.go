package contract

import "errors"

// Errors of the library (docs/contrato-plantillas.md §13). Every package of
// the module returns these, wrapped with context, so callers can use
// errors.Is without knowing which package produced them.
var (
	// ErrOffline: there is no network and nothing usable in the cache.
	ErrOffline = errors.New("sin conexión y sin copia en caché")
	// ErrNotFound: the template, version or pack is not in the catalog.
	ErrNotFound = errors.New("no existe en el catálogo")
	// ErrUnsupportedSchema: an index or manifest uses an unknown schemaVersion.
	ErrUnsupportedSchema = errors.New("versión del contrato no soportada")
	// ErrChecksum: the downloaded bundle does not match the index's sha256.
	ErrChecksum = errors.New("el sha256 del bundle no coincide")
	// ErrTooLarge: a size, count or compression-ratio limit was exceeded.
	ErrTooLarge = errors.New("se superó un límite de tamaño")
	// ErrUnsafePath: a bundle entry breaks rule B2 or B3 (path or symlink).
	ErrUnsafePath = errors.New("ruta insegura en el bundle")
	// ErrForbiddenFile: a bundle entry breaks rule B4 (file not allowed).
	ErrForbiddenFile = errors.New("archivo no permitido en el bundle")
	// ErrBundleLayout: the bundle breaks rule B1, B6 or B7.
	ErrBundleLayout = errors.New("estructura de bundle inválida")
	// ErrDestExists: the destination exists and is not empty.
	ErrDestExists = errors.New("el destino ya existe y no está vacío")
)
