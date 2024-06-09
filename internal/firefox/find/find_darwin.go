//go:build darwin && !ios

package find

import (
	"os"
	"path/filepath"
)

func firefoxRoots(yield func(string, error) bool) {
	// "$HOME/Library/Application Support"
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		_ = yield(``, err)
		return
	}
	if !yield(filepath.Join(cfgDir, `Firefox`), nil) {
		return
	}
}
