//go:build darwin && !ios

package edge

import (
	"os"
	"path/filepath"
)

func edgeChromiumRoots(yield func(string, error) bool) {
	// "$HOME/Library/Application Support"
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		_ = yield(``, err)
		return
	}
	if !yield(filepath.Join(cfgDir, `Microsoft Edge`), nil) {
		return
	}
}
