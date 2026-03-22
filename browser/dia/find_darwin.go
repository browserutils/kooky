//go:build darwin && !ios

package dia

import (
	"os"
	"path/filepath"
)

func diaChromiumRoots(yield func(string, error) bool) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		_ = yield(``, err)
		return
	}
	_ = yield(filepath.Join(cfgDir, `Dia`, `User Data`), nil)
}
