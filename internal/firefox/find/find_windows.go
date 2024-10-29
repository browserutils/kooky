//go:build windows
// +build windows

package find

import (
	"os"
	"path/filepath"
)

func firefoxRoots(yield func(string, error) bool) {
	// "%AppData%"
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		_ = yield(``, err)
		return
	}
	if !yield(filepath.Join(cfgDir, `Mozilla`, `Firefox`), nil) {
		return
	}
}
