//go:build linux && !android

package vivaldi

import (
	"os"
	"path/filepath"

	"github.com/browserutils/kooky/internal/windowsx"
)

func vivaldiRoots(yield func(string, error) bool) {
	// "${XDG_CONFIG_HOME:-$HOME/.config}"
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		_ = yield(``, err)
		return
	}
	for _, dir := range []string{`vivaldi`, `vivaldi-snapshot`} {
		if !yield(filepath.Join(cfgDir, dir), nil) {
			return
		}
	}
	// on WSL Linux add Windows paths
	if !windowsx.IsWSL() {
		return
	}
	for r, err := range windowsVivaldiRoots {
		if !yield(r, err) {
			return
		}
	}
}
