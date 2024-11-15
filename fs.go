package logftxt

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pamburus/logftxt/internal/pkg/cfgdir"
	"github.com/pamburus/logftxt/internal/pkg/pathx"
)

// SystemFS constructs default FS implementation that accesses operating system's root filesystem.
func SystemFS() FS {
	return &systemFS{}
}

// FS is an abstract filesystem for accessing configuration files.
type FS interface {
	ConfigDir() (fs.FS, error)
	Open(filename string) (fs.File, error)
}

// WithFS returns an FSOption.
func WithFS(value FS) FSOption {
	return FSOption{value}
}

// ---

// FSOption is an optional parameter created by WithFS that can be used in
//   - LoadConfig
//   - ConfigFromEnvironment
//   - ConfigFromDefaultPath
//   - LoadTheme
//   - ThemeFromEnvironment.
type FSOption struct {
	value FS
}

func (v FSOption) toAppenderOptions(o *appenderOptions) {
	o.fs = v.value
}

func (v FSOption) toEncoderOptions(o *encoderOptions) {
	o.fs = v.value
}

func (v FSOption) toDomain(d *domain) {
	d.fs = v.value
}

// ---

type systemFS struct{}

func (f *systemFS) ConfigDir() (fs.FS, error) {
	dir, err := cfgdir.Locate()
	if err != nil {
		return nil, fmt.Errorf("failed to locate configuration directory: %w", err)
	}

	return os.DirFS(dir), nil
}

func (f *systemFS) Open(filename string) (fs.File, error) {
	if filepath.IsAbs(filename) || pathx.OS().HasPrefix(filename, ".") || pathx.OS().HasPrefix(filename, "..") {
		result, err := os.Open(filename) //nolint:gosec // it is ok to open theme files requested by the user
		if err != nil {
			return nil, fmt.Errorf("os: %w", err)
		}

		return result, nil
	}

	configDir, err := f.ConfigDir()
	if err != nil {
		return nil, err
	}

	result, err := configDir.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("config dir: %w", err)
	}

	return result, nil
}
