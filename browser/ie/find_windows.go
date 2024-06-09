//go:build windows

package ie

import (
	"os"
	"path/filepath"
)

func ieRoots(yield func(string, error) bool) {
	confDir, err := os.UserConfigDir()
	if err != nil {
		_ = yield(``, err)
		return
	}

	if !yield(filepath.Join(confDir, `Microsoft`, `Windows`, `Cookies`), nil) {
		return
	}
	if !yield(filepath.Join(confDir, `Microsoft`, `Windows`, `Cookies`, `Low`), nil) {
		return
	}
}
