//go:build windows || linux

package edge

import (
	"path/filepath"

	"github.com/browserutils/kooky/internal/windowsx"
)

// for Windows and WSL

func windowsEdgeRoots(yield func(string, error) bool) {
	// %LocalAppData%
	locApp, err := windowsx.LocalAppData()
	if err != nil {
		_ = yield(``, err)
		return
	}
	_ = yield(filepath.Join(locApp, `Microsoft`, `Edge`, `User Data`), nil)
}
