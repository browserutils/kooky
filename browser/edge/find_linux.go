//go:build linux && !android

package edge

import (
	"os"
	"path/filepath"

	"github.com/browserutils/kooky/internal/windowsx"
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
	// on WSL Linux add Windows paths
	if !windowsx.IsWSL() {
		return
	}
	for r, err := range windowsEdgeRoots {
		if !yield(r, err) {
			return
		}
	}
}
