package logftxt

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pamburus/logftxt/internal/pkg/cfgdir"
	"github.com/pamburus/logftxt/internal/pkg/pathx"
)

// DefaultFS constructs default FS implementation.
func DefaultFS() FS {
	return &defaultFS{}
}

// FS is an abstract filesystem for accessing configuration files.
type FS interface {
	ConfigDir() (fs.FS, error)
	Open(filename string) (fs.File, error)
}

// ---

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

func (v FSOption) toFSOptions(o *fsOptions) {
	o.fs = v.value
}

func (v FSOption) toFSEnvOptions(o *fsEnvOptions) {
	o.fs = v.value
}

// ---

type defaultFS struct{}

func (f *defaultFS) ConfigDir() (fs.FS, error) {
	dir, err := cfgdir.Locate()
	if err != nil {
		return nil, fmt.Errorf("failed to locate configuration directory: %v", err)
	}

	return os.DirFS(dir), nil
}

func (f *defaultFS) Open(filename string) (fs.File, error) {
	if filepath.IsAbs(filename) || pathx.OS().HasPrefix(filename, ".") || pathx.OS().HasPrefix(filename, "..") {
		return os.Open(filename) //nolint:gosec
	}

	configDir, err := f.ConfigDir()
	if err != nil {
		return nil, err
	}

	return configDir.Open(filename)
}

// ---

type fsOption interface {
	toFSOptions(*fsOptions)
	fsEnvOption
}

// ---

type fsOptions struct {
	fs FS
}

func (o fsOptions) With(other []fsOption) fsOptions {
	for _, oo := range other {
		oo.toFSOptions(&o)
	}

	return o
}

func (o fsOptions) WithDefaults() fsOptions {
	if o.fs == nil {
		o.fs = DefaultFS()
	}

	return o
}
