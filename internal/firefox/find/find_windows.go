//+build windows

package find

import (
	"os"
	"path/filepath"
)

func firefoxRoots() ([]string, error) {
	// "%AppData%"
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return []string{filepath.Join(cfgDir, `Mozilla`, `Firefox`)}, nil
}
