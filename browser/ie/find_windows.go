//go:build windows

package ie

import (
	"path/filepath"

	"github.com/browserutils/kooky/internal/windowsx"
)

func ieRoots(yield func(string, error) bool) {
	appData, err := windowsx.AppData()
	if err != nil {
		_ = yield(``, err)
		return
	}

	if !yield(filepath.Join(appData, `Microsoft`, `Windows`, `Cookies`), nil) {
		return
	}
	if !yield(filepath.Join(appData, `Microsoft`, `Windows`, `Cookies`, `Low`), nil) {
		return
	}
}
