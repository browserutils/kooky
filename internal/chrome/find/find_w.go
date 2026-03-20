//go:build windows || linux

package find

import (
	"path/filepath"

	"github.com/browserutils/kooky/internal/windowsx"
)

// for Windows and WSL

func windowsChromeRoots(yield func(string, error) bool) {
	locApp, err := windowsx.LocalAppData()
	if err != nil {
		_ = yield(``, err)
		return
	}
	// https://chromium.googlesource.com/chromium/src.git/+/62.0.3202.58/docs/user_data_dir.md#windows
	if !yield(filepath.Join(locApp, `Google`, `Chrome`, `User Data`), nil) {
		return
	}
	// Canary uses InstallConstants::install_suffix
	// https://cs.chromium.org/chromium/src/chrome/install_static/install_constants.h?q=install_suffix
	if !yield(filepath.Join(locApp, `Google`, `Chrome SxS`, `User Data`), nil) {
		return
	}
}

var (
	windowsChromiumRoots = locAppRoots(`Chromium`, `User Data`)
	windowsBraveRoots    = locAppRoots(`BraveSoftware`, `Brave-Browser`, `User Data`)
)

func locAppRoots(pathParts ...string) func(yield func(string, error) bool) {
	return func(yield func(string, error) bool) {
		locApp, err := windowsx.LocalAppData()
		if err != nil {
			_ = yield(``, err)
			return
		}
		if !yield(filepath.Join(append([]string{locApp}, pathParts...)...), nil) {
			return
		}
	}
}
