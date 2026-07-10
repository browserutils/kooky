//go:build darwin && !ios

package vivaldi

import (
	"os"
	"path/filepath"
)

func vivaldiRoots(yield func(string, error) bool) {
	// "$HOME/Library/Application Support"
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		_ = yield(``, err)
		return
	}
	if !yield(filepath.Join(cfgDir, `Vivaldi`), nil) {
		return
	}
}
