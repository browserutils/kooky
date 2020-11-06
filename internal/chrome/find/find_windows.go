//+build windows

package find

import (
	"errors"
	"os"
	"path/filepath"
)

func chromeRoots() ([]string, error) {
	// https://chromium.googlesource.com/chromium/src.git/+/62.0.3202.58/docs/user_data_dir.md#windows
	cfgDir := os.Getenv(`LocalAppData`)
	if len(cfgDir) == 0 {
		return nil, errors.New(`%LocalAppData% is empty`)
	}
	var ret = []string{
		filepath.Join(cfgDir, `Google`, `Chrome`, `User Data`),
		// Canary uses InstallConstants::install_suffix
		// https://cs.chromium.org/chromium/src/chrome/install_static/install_constants.h?q=install_suffix
		filepath.Join(cfgDir, `Google`, `Chrome SxS`, `User Data`),
	}
	return ret, nil
}

func chromiumRoots() ([]string, error) {
	cfgDir := os.Getenv(`LocalAppData`)
	if len(cfgDir) == 0 {
		return nil, errors.New(`%LocalAppData% is empty`)
	}
	return []string{filepath.Join(cfgDir, `Chromium`, `User Data`)}, nil
}
