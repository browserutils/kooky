//go:build darwin

package find

import (
	"os"
	"path/filepath"
)

func edgeRoots() ([]string, error) {
	// "$HOME/Library/Application Support"
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return []string{filepath.Join(cfgDir, `Microsoft Edge`)}, nil
}
