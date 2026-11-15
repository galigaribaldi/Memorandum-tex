package bundle

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/galigaribaldi/Memorandum-tex/tools/pkg/contract"
)

// Limits of SafeExtract (B5).
type Limits struct {
	MaxFiles int
	MaxBytes int64 // total, uncompressed
	// MaxRatio is the largest uncompressed/compressed ratio of a file bigger
	// than RatioFrom bytes. Small files are exempt: 1 KB of blanks
	// compresses far beyond 100:1 and is not an attack.
	MaxRatio  int64
	RatioFrom int64
}

// DefaultLimits are those of the contract, the same as Creative-Ink's copyTemplate.
var DefaultLimits = Limits{
	MaxFiles:  contract.MaxBundleFiles,
	MaxBytes:  contract.MaxBundleBytes,
	MaxRatio:  contract.MaxCompressionRatio,
	RatioFrom: 1 << 20,
}

// ExtractOptions of SafeExtract.
type ExtractOptions struct {
	// ExpectID and ExpectVersion, when set, must match the bundle's manifest
	// (what the index said was being downloaded).
	ExpectID, ExpectVersion string
	// Limits; the zero value means DefaultLimits.
	Limits Limits
}

// Result describes an extracted bundle.
type Result struct {
	ID, Version string
	Manifest    *contract.Manifest
	Dir         string   // = dest
	MainFile    string   // entry, relative to Dir with "/"
	Files       []string // every file, relative to Dir with "/", sorted
}

// SHA256 returns the lowercase hex sha256 of r and how many bytes it read.
func SHA256(r io.Reader) (string, int64, error) {
	h := sha256.New()
	n, err := io.Copy(h, r)
	return hex.EncodeToString(h.Sum(nil)), n, err
}

// VerifySHA256 checks the whole zip against the index's hash (B9). It must
// run before SafeExtract.
func VerifySHA256(r io.Reader, want string) error {
	got, _, err := SHA256(r)
	if err != nil {
		return err
	}
	if !contract.ValidSHA256(want) || got != want {
		return fmt.Errorf("%w: se esperaba %s, se obtuvo %s", contract.ErrChecksum, want, got)
	}
	return nil
}

// SafeExtractFile opens zipPath and calls SafeExtract.
func SafeExtractFile(ctx context.Context, zipPath, dest string, opt ExtractOptions) (*Result, error) {
	f, err := os.Open(zipPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return SafeExtract(ctx, f, info.Size(), dest, opt)
}

// SafeExtract checks the whole bundle (B1–B7, B10) and only then extracts it
// into dest, without its root folder. Files go first to a temporary folder
// next to dest that is renamed at the end, so a failure or a cancellation
// never leaves a half-extracted template. dest must not exist or be empty.
func SafeExtract(ctx context.Context, r io.ReaderAt, size int64, dest string, opt ExtractOptions) (*Result, error) {
	if opt.Limits == (Limits{}) {
		opt.Limits = DefaultLimits
	}
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("%w: no es un zip válido: %v", contract.ErrBundleLayout, err)
	}
	plan, err := check(zr, opt)
	if err != nil {
		return nil, err
	}
	if err := emptyOrMissing(dest); err != nil {
		return nil, err
	}

	parent := filepath.Dir(dest)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return nil, err
	}
	tmp, err := os.MkdirTemp(parent, ".ink-extract-*")
	if err != nil {
		return nil, err
	}
	done := false
	defer func() {
		if !done {
			_ = os.RemoveAll(tmp) // our own temporary folder, created above
		}
	}()
	for _, f := range plan.files {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := writeEntry(f.zf, filepath.Join(tmp, filepath.FromSlash(f.rel))); err != nil {
			return nil, err
		}
	}
	if err := os.Remove(dest); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err // dest existed and was empty: rename needs it gone (Windows)
	}
	if err := os.Rename(tmp, dest); err != nil {
		return nil, err
	}
	done = true

	res := &Result{ID: plan.m.ID, Version: plan.m.Version, Manifest: plan.m, Dir: dest, MainFile: plan.m.Entry}
	for _, f := range plan.files {
		res.Files = append(res.Files, f.rel)
	}
	return res, nil
}

type plannedFile struct {
	zf  *zip.File
	rel string // path without the root folder
}

type plan struct {
	m     *contract.Manifest
	files []plannedFile
}

// check validates every entry and the layout without writing anything.
func check(zr *zip.Reader, opt ExtractOptions) (*plan, error) {
	lim := opt.Limits
	var (
		root  string
		files []plannedFile
		total int64
		seen  = map[string]string{}
		byRel = map[string]*zip.File{}
	)
	for _, zf := range zr.File {
		name := zf.Name
		// B2: nothing absolute or with Windows separators, on any OS.
		if strings.HasPrefix(name, "/") || strings.Contains(name, `\`) {
			return nil, fmt.Errorf("%w: %q", contract.ErrUnsafePath, name)
		}
		top, rest, _ := strings.Cut(strings.TrimSuffix(name, "/"), "/")
		switch {
		case root == "":
			root = top
		case top != root: // B1
			return nil, fmt.Errorf("%w: más de una carpeta raíz (%s y %s)", contract.ErrBundleLayout, root, top)
		}
		if rest == "" {
			continue // the root folder itself
		}
		if !contract.ValidRelPath(rest) || !filepath.IsLocal(filepath.FromSlash(rest)) {
			return nil, fmt.Errorf("%w: %q", contract.ErrUnsafePath, name)
		}
		mode := zf.Mode()
		if mode&fs.ModeSymlink != 0 { // B3
			return nil, fmt.Errorf("%w: %q es un enlace simbólico", contract.ErrUnsafePath, name)
		}
		if zf.FileInfo().IsDir() {
			continue
		}
		if !mode.IsRegular() {
			return nil, fmt.Errorf("%w: %q no es un archivo normal", contract.ErrUnsafePath, name)
		}
		if base := path.Base(rest); !contract.AllowedBundleFile(base) { // B4
			return nil, fmt.Errorf("%w: %q", contract.ErrForbiddenFile, name)
		}
		lower := strings.ToLower(rest) // B10
		if first, dup := seen[lower]; dup {
			return nil, fmt.Errorf("%w: %q y %q solo difieren en mayúsculas", contract.ErrUnsafePath, first, rest)
		}
		seen[lower] = rest

		size := int64(zf.UncompressedSize64) // B5, with the declared size; writeEntry enforces it
		if size < 0 || size > lim.MaxBytes {
			return nil, fmt.Errorf("%w: %q", contract.ErrTooLarge, name)
		}
		if size >= lim.RatioFrom && (zf.CompressedSize64 == 0 || size/int64(zf.CompressedSize64) > lim.MaxRatio) {
			return nil, fmt.Errorf("%w: %q se comprime más de %d:1 (¿bomba zip?)", contract.ErrTooLarge, name, lim.MaxRatio)
		}
		if total += size; total > lim.MaxBytes {
			return nil, fmt.Errorf("%w: el bundle supera %d MB descomprimido", contract.ErrTooLarge, lim.MaxBytes>>20)
		}
		if len(files) >= lim.MaxFiles {
			return nil, fmt.Errorf("%w: más de %d archivos", contract.ErrTooLarge, lim.MaxFiles)
		}
		files = append(files, plannedFile{zf: zf, rel: rest})
		byRel[rest] = zf
	}
	if root == "" {
		return nil, fmt.Errorf("%w: el zip está vacío", contract.ErrBundleLayout)
	}

	// B6: the required files, and a manifest that matches the root and the index.
	for _, req := range []string{"template.json", contract.LicenseFileName, "README.md", contract.MetaFileName} {
		if byRel[req] == nil {
			return nil, fmt.Errorf("%w: falta %s", contract.ErrBundleLayout, req)
		}
	}
	tj, err := readSmall(byRel["template.json"], contract.MaxManifestBytes)
	if err != nil {
		return nil, err
	}
	m, rep := contract.ValidateManifest(tj, contract.Options{Origin: contract.OriginRepository})
	switch {
	case m == nil || rep.HasErrors():
		return nil, fmt.Errorf("%w: template.json inválido: %v", contract.ErrBundleLayout, rep.Errors())
	case m.ID != root:
		return nil, fmt.Errorf("%w: la carpeta raíz es %q pero el manifiesto dice %q", contract.ErrBundleLayout, root, m.ID)
	case opt.ExpectID != "" && m.ID != opt.ExpectID:
		return nil, fmt.Errorf("%w: se esperaba la plantilla %q y llegó %q", contract.ErrBundleLayout, opt.ExpectID, m.ID)
	case opt.ExpectVersion != "" && m.Version != opt.ExpectVersion:
		return nil, fmt.Errorf("%w: se esperaba la versión %s y llegó %s", contract.ErrBundleLayout, opt.ExpectVersion, m.Version)
	case byRel[m.Entry] == nil:
		return nil, fmt.Errorf("%w: falta el archivo principal %s", contract.ErrBundleLayout, m.Entry)
	case m.Build.ShellEscape: // defense in depth: compat blocks it before downloading
		return nil, fmt.Errorf("%w: la plantilla pide --shell-escape, prohibido en v1", contract.ErrForbiddenFile)
	}

	// B7: .creative-ink-meta is exactly {version: 1, mainFile: entry}.
	raw, err := readSmall(byRel[contract.MetaFileName], 4<<10)
	if err != nil {
		return nil, err
	}
	var meta map[string]json.RawMessage
	var version int
	var mainFile string
	if json.Unmarshal(raw, &meta) != nil || len(meta) != 2 ||
		json.Unmarshal(meta["version"], &version) != nil || json.Unmarshal(meta["mainFile"], &mainFile) != nil ||
		version != 1 || mainFile != m.Entry {
		return nil, fmt.Errorf("%w: %s debe ser {\"version\": 1, \"mainFile\": %q}", contract.ErrBundleLayout, contract.MetaFileName, m.Entry)
	}

	sort.Slice(files, func(i, j int) bool { return files[i].rel < files[j].rel })
	return &plan{m: m, files: files}, nil
}

// writeEntry extracts one file, creating it exclusively and reading at most
// its declared size: a zip that lies about a size fails instead of filling the disk.
func writeEntry(zf *zip.File, target string) error {
	rc, err := zf.Open()
	if err != nil {
		return fmt.Errorf("%w: %s: %v", contract.ErrBundleLayout, zf.Name, err)
	}
	defer rc.Close()
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	declared := int64(zf.UncompressedSize64)
	n, cerr := io.Copy(out, io.LimitReader(rc, declared+1))
	if err := out.Close(); cerr == nil {
		cerr = err
	}
	switch {
	case n > declared:
		return fmt.Errorf("%w: %s ocupa más de lo que declara el zip", contract.ErrTooLarge, zf.Name)
	case cerr != nil:
		return fmt.Errorf("%w: %s: %v", contract.ErrBundleLayout, zf.Name, cerr)
	}
	return nil
}

func readSmall(zf *zip.File, limit int64) ([]byte, error) {
	if int64(zf.UncompressedSize64) > limit {
		return nil, fmt.Errorf("%w: %s", contract.ErrTooLarge, zf.Name)
	}
	rc, err := zf.Open()
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", contract.ErrBundleLayout, zf.Name, err)
	}
	defer rc.Close()
	var b bytes.Buffer
	if _, err := io.Copy(&b, io.LimitReader(rc, limit+1)); err != nil {
		return nil, fmt.Errorf("%w: %s: %v", contract.ErrBundleLayout, zf.Name, err)
	}
	if int64(b.Len()) > limit {
		return nil, fmt.Errorf("%w: %s", contract.ErrTooLarge, zf.Name)
	}
	return b.Bytes(), nil
}

// emptyOrMissing fails with ErrDestExists if dest has anything in it.
func emptyOrMissing(dest string) error {
	entries, err := os.ReadDir(dest)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return fmt.Errorf("%w: %s: %v", contract.ErrDestExists, dest, err)
	case len(entries) > 0:
		return fmt.Errorf("%w: %s", contract.ErrDestExists, dest)
	}
	return nil
}
