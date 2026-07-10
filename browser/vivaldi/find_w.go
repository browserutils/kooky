//go:build windows || linux

package vivaldi

import (
	"path/filepath"

	"github.com/browserutils/kooky/internal/windowsx"
)

// for Windows and WSL

func windowsVivaldiRoots(yield func(string, error) bool) {
	// %LocalAppData%
	locApp, err := windowsx.LocalAppData()
	if err != nil {
		_ = yield(``, err)
		return
	}
	_ = yield(filepath.Join(locApp, `Vivaldi`, `User Data`), nil)
}
