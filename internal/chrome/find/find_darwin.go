//go:build darwin && !ios

package find

import (
	"os"
	"path/filepath"
)

func chromeRoots(yield func(string, error) bool) {
	// https://chromium.googlesource.com/chromium/src.git/+/62.0.3202.58/docs/user_data_dir.md#mac-os-x
	// The canary channel suffix is determined using the CrProductDirName key in the browser app's Info.plist
	// "$HOME/Library/Application Support"
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		_ = yield(``, err)
		return
	}
	if !yield(filepath.Join(cfgDir, `Google`, `Chrome`), nil) {
		return
	}
	if !yield(filepath.Join(cfgDir, `Google`, `Chrome Canary`), nil) {
		return
	}
}

func chromiumRoots(yield func(string, error) bool) {
	// "$HOME/Library/Application Support"
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		_ = yield(``, err)
		return
	}
	if !yield(filepath.Join(cfgDir, `Chromium`), nil) {
		return
	}
}
