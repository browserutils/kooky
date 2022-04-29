//go:build darwin && !ios

package find

import (
	"os"
	"path/filepath"
)

func chromeRoots() ([]string, error) {
	// https://chromium.googlesource.com/chromium/src.git/+/62.0.3202.58/docs/user_data_dir.md#mac-os-x
	// The canary channel suffix is determined using the CrProductDirName key in the browser app's Info.plist
	// "$HOME/Library/Application Support"
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	var ret = []string{
		filepath.Join(cfgDir, `Google`, `Chrome`),
		filepath.Join(cfgDir, `Google`, `Chrome Canary`),
	}
	return ret, nil
}

func chromiumRoots() ([]string, error) {
	// "$HOME/Library/Application Support"
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	var ret = []string{
		filepath.Join(cfgDir, `Chromium`),
	}
	return ret, nil
}
