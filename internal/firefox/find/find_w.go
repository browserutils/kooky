//go:build windows || linux

package find

import (
	"path/filepath"

	"github.com/browserutils/kooky/internal/windowsx"
)

// for Windows and WSL

func windowsFirefoxRoots(yield func(string, error) bool) {
	// "%AppData%"
	appData, err := windowsx.AppData()
	if err != nil {
		_ = yield(``, err)
		return
	}
	_ = yield(filepath.Join(appData, `Mozilla`, `Firefox`), nil)
}
