//go:build linux && !android

package edge

import (
	"os"
	"path/filepath"
)

func edgeChromiumRoots(yield func(string, error) bool) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		_ = yield(``, err)
		return
	}
	if !yield(filepath.Join(cfgDir, `microsoft-edge`), nil) {
		return
	}
}
