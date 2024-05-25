//go:build linux && !android

package edge

import (
	"os"
	"path/filepath"
)

func edgeChromiumRoots() ([]string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return []string{filepath.Join(cfgDir, `microsoft-edge`)}, nil
}
