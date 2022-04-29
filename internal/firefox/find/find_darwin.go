//go:build darwin && !ios

package find

import (
	"os"
	"path/filepath"
)

func firefoxRoots() ([]string, error) {
	// "$HOME/Library/Application Support"
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return []string{filepath.Join(cfgDir, `Firefox`)}, nil
}
