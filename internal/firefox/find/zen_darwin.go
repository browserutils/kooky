//go:build darwin && !ios

package find

import (
	"os"
	"path/filepath"
)

func zenRoots(yield func(string, error) bool) {
	// "$HOME/Library/Application Support/zen"
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		_ = yield(``, err)
		return
	}
	if !yield(filepath.Join(cfgDir, `zen`), nil) {
		return
	}
}
