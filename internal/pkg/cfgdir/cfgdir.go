// Package cfgdir provides helpers for dealing with configuration directories.
package cfgdir

import (
	"os"
	"path/filepath"
)

// Locate returns path to the configuration directory.
func Locate() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".config", "logftxt"), nil
}
