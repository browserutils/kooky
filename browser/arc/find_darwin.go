//go:build darwin && !ios

package arc

import (
	"os"
	"path/filepath"
)

func arcChromiumRoots(yield func(string, error) bool) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		_ = yield(``, err)
		return
	}
	_ = yield(filepath.Join(cfgDir, `Arc`, `User Data`), nil)
}
