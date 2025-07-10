//go:build windows || linux

package find

import (
	"errors"
	"path/filepath"
)

// for Windows and WSL

func windowsChromeRoots(dir string) func(yield func(string, error) bool) {
	return func(yield func(string, error) bool) {
		// https://chromium.googlesource.com/chromium/src.git/+/62.0.3202.58/docs/user_data_dir.md#windows
		if len(dir) == 0 {
			_ = yield(``, errors.New(`%LocalAppData% is empty`))
			return
		}
		if !yield(filepath.Join(dir, `Google`, `Chrome`, `User Data`), nil) {
			return
		}
		// Canary uses InstallConstants::install_suffix
		// https://cs.chromium.org/chromium/src/chrome/install_static/install_constants.h?q=install_suffix
		if !yield(filepath.Join(dir, `Google`, `Chrome SxS`, `User Data`), nil) {
			return
		}
	}
}

func windowsChromiumRoots(dir string) func(yield func(string, error) bool) {
	return func(yield func(string, error) bool) {
		if len(dir) == 0 {
			_ = yield(``, errors.New(`%LocalAppData% is empty`))
			return
		}
		if !yield(filepath.Join(dir, `Chromium`, `User Data`), nil) {
			return
		}
	}
}
