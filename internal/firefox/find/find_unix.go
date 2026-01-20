//go:build !windows && !darwin && !plan9 && !android && !js && !aix

package find

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/browserutils/kooky/internal/wsl"
)

func firefoxRoots(yield func(string, error) bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		_ = yield(``, err)
		return
	}
	// Ubuntu 21.10 (snap)
	if !yield(filepath.Join(home, `snap`, `firefox`, `common`, `.mozilla`, `firefox`), nil) {
		return
	}
	if !yield(filepath.Join(home, `.mozilla`, `firefox`), nil) {
		return
	}
	// Mozilla PPA
	if !yield(filepath.Join(home, `.mozilla`, `firefox-esr`), nil) {
		return
	}
	// on WSL Linux add Windows paths
	if runtime.GOOS != `linux` {
		return
	}
	appDataRoot, err := wsl.WSLAppDataRoot()
	if err != nil && (errors.Is(err, wsl.ErrNotWSL) || !yield(``, err)) {
		return
	}
	if !yield(filepath.Join(appDataRoot, `Roaming`, `Mozilla`, `Firefox`), nil) {
		return
	}
}
