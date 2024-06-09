//go:build windows
// +build windows

package find

import (
	"errors"
	"os"
	"path/filepath"
)

func chromeRoots(yield func(string, error) bool) {
	// https://chromium.googlesource.com/chromium/src.git/+/62.0.3202.58/docs/user_data_dir.md#windows
	cfgDir := os.Getenv(`LocalAppData`)
	if len(cfgDir) == 0 {
		_ = yield(``, errors.New(`%LocalAppData% is empty`))
		return
	}
	if !yield(filepath.Join(cfgDir, `Google`, `Chrome`, `User Data`), nil) {
		return
	}
	// Canary uses InstallConstants::install_suffix
	// https://cs.chromium.org/chromium/src/chrome/install_static/install_constants.h?q=install_suffix
	if !yield(filepath.Join(cfgDir, `Google`, `Chrome SxS`, `User Data`), nil) {
		return
	}
}

func chromiumRoots(yield func(string, error) bool) {
	cfgDir := os.Getenv(`LocalAppData`)
	if len(cfgDir) == 0 {
		_ = yield(``, errors.New(`%LocalAppData% is empty`))
		return
	}
	if !yield(filepath.Join(cfgDir, `Chromium`, `User Data`), nil) {
		return
	}
}
