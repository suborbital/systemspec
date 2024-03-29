package bundle

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/suborbital/systemspec/tenant"
)

// Bundle represents a Module bundle.
type Bundle struct {
	filepath     string
	TenantConfig *tenant.Config
	staticFiles  map[string]bool
}

// StaticFile returns a static file from the bundle, if it exists.
func (b *Bundle) StaticFile(filePathIn string) ([]byte, error) {
	// normalize in case the caller added `/` or `./` to the filename.
	filePath := NormalizeStaticFilename(filePathIn)

	if _, exists := b.staticFiles[filePath]; !exists {
		return nil, os.ErrNotExist
	}

	r, err := zip.OpenReader(b.filepath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open bundle")
	}

	defer r.Close()

	// re-add the static/ prefix to ensure sandboxing to the static directory.
	staticFilePath := ensurePrefix(filePath, "static/")

	var contents []byte

	var zipFile *zip.File

	for _, f := range r.File {
		if f.Name == staticFilePath {
			zipFile = f
			break
		}
	}

	file, err := zipFile.Open()
	if err != nil {
		return nil, errors.Wrap(err, "failed to Open static file")
	}

	defer func() {
		_ = file.Close()
	}()

	contents, err = io.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ReadAll static file")
	}

	return contents, nil
}

// Write writes a module bundle
// based loosely on https://golang.org/src/archive/zip/example_test.go
// staticFiles should be a map of *relative* filepaths to their associated files, with or without the `static/` prefix.
func Write(tenantConfigBytes []byte, modules []os.File, staticFiles map[string]os.File, targetPath string) error {
	if len(tenantConfigBytes) == 0 {
		return errors.New("tenant config must be provided")
	}

	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive.
	w := zip.NewWriter(buf)

	// Add tenant config to archive.
	if err := writeTenantConfig(w, tenantConfigBytes); err != nil {
		return errors.Wrap(err, "failed to writeTenantConfig")
	}

	// Add the Wasm modules to the archive.
	for _, file := range modules {
		file := file

		if file.Name() == "tenant.json" {
			// only allow the canonical tenant config that's passed in.
			continue
		}

		contents, err := io.ReadAll(&file)
		if err != nil {
			return errors.Wrapf(err, "failed to read file %s", file.Name())
		}

		if err := writeFile(w, filepath.Base(file.Name()), contents); err != nil {
			return errors.Wrap(err, "failed to writeFile into bundle")
		}
	}

	// Add static files to the archive.
	for path, file := range staticFiles {
		file := file

		contents, err := io.ReadAll(&file)
		if err != nil {
			return errors.Wrapf(err, "failed to read file %s", file.Name())
		}

		fileName := ensurePrefix(path, "static/")
		if err := writeFile(w, fileName, contents); err != nil {
			return errors.Wrap(err, "failed to writeFile into bundle")
		}
	}

	if err := w.Close(); err != nil {
		return errors.Wrap(err, "failed to close bundle writer")
	}

	if err := os.WriteFile(targetPath, buf.Bytes(), 0600); err != nil {
		return errors.Wrap(err, "failed to write bundle to disk")
	}

	return nil
}

func writeTenantConfig(w *zip.Writer, tenantConfigBytes []byte) error {
	if err := writeFile(w, "tenant.json", tenantConfigBytes); err != nil {
		return errors.Wrap(err, "failed to writeFile for tenant.json")
	}

	return nil
}

func writeFile(w *zip.Writer, name string, contents []byte) error {
	f, err := w.Create(name)
	if err != nil {
		return errors.Wrap(err, "failed to add file to bundle")
	}

	_, err = f.Write(contents)
	if err != nil {
		return errors.Wrap(err, "failed to write file into bundle")
	}

	return nil
}

// Read reads a .wasm.zip file and returns the bundle of wasm modules
// (suitable to be loaded into a wasmer instance).
func Read(path string) (*Bundle, error) {
	// Open a zip archive for reading.
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open bundle")
	}

	defer r.Close()

	bundle := &Bundle{
		filepath:    path,
		staticFiles: map[string]bool{},
	}

	// first, find the tenant config.
	for _, f := range r.File {
		if f.Name == "tenant.json" {
			tenantConfig, err := readTenantConfig(f)
			if err != nil {
				return nil, errors.Wrap(err, "failed to readTenantConfig from bundle")
			}

			bundle.TenantConfig = tenantConfig

			continue
		}
	}

	if bundle.TenantConfig == nil {
		return nil, errors.New("bundle is missing tenant.json")
	}

	// Iterate through the files in the archive.
	for _, f := range r.File {
		if f.Name == "tenant.json" {
			// we already have a tenant config by now.
			continue
		} else if strings.HasPrefix(f.Name, "static/") {
			// build up the list of available static files in the bundle for quick reference later.
			filePath := strings.TrimPrefix(f.Name, "static/")
			bundle.staticFiles[filePath] = true
			continue
		} else if !strings.HasSuffix(f.Name, ".wasm") {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open %s from bundle", f.Name)
		}

		wasmBytes, err := io.ReadAll(rc)
		if err != nil {
			_ = rc.Close()
			return nil, errors.Wrapf(err, "failed to read %s from bundle", f.Name)
		}

		_ = rc.Close()

		// for now, the bundle spec only supports the default namespace
		FQMN := fmt.Sprintf("/name/default/%s", strings.TrimSuffix(f.Name, ".wasm"))

		module, err := bundle.TenantConfig.FindModule(FQMN)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to FindModule for %s (%s)", f.Name, FQMN)
		} else if module == nil {
			return nil, fmt.Errorf("unable to find Module for Wasm file %s (%s)", f.Name, FQMN)
		}

		module.WasmRef = tenant.NewWasmModuleRef(f.Name, module.FQMN, wasmBytes)
	}

	if bundle.TenantConfig == nil {
		return nil, errors.New("bundle did not contain tenantConfig")
	}

	return bundle, nil
}

func readTenantConfig(f *zip.File) (*tenant.Config, error) {
	file, err := f.Open()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open %s from bundle", f.Name)
	}

	tenantConfigBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s from bundle", f.Name)
	}

	d := &tenant.Config{}
	if err := d.Unmarshal(tenantConfigBytes); err != nil {
		return nil, errors.Wrap(err, "failed to Unmarshal tenant config")
	}

	return d, nil
}

func ensurePrefix(val, prefix string) string {
	if strings.HasPrefix(val, prefix) {
		return val
	}

	return fmt.Sprintf("%s%s", prefix, val)
}

// NormalizeStaticFilename will take various variations of a filename and
// normalize it to what is listed in the staticFile name cache on the Bundle struct.
func NormalizeStaticFilename(fileName string) string {
	withoutStatic := strings.TrimPrefix(fileName, "static/")
	withoutLeadingSlash := strings.TrimPrefix(withoutStatic, "/")
	withoutDotSlash := strings.TrimPrefix(withoutLeadingSlash, "./")

	return withoutDotSlash
}
