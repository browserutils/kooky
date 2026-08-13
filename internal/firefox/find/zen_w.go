//go:build windows || linux

package find

import (
	"path/filepath"

	"github.com/browserutils/kooky/internal/windowsx"
)

// for Windows and WSL

func windowsZenRoots(yield func(string, error) bool) {
	// "%AppData%\zen" — https://docs.zen-browser.app/faq
	appData, err := windowsx.AppData()
	if err != nil {
		_ = yield(``, err)
		return
	}
	_ = yield(filepath.Join(appData, `zen`), nil)
}
