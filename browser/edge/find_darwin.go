//go:build darwin && !ios

package edge

import (
	"os"
	"path/filepath"
)

func edgeChromiumRoots() ([]string, error) {
	// "$HOME/Library/Application Support"
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return []string{filepath.Join(cfgDir, `Microsoft Edge`)}, nil
}
