//go:build !windows && !darwin && !plan9 && !android && !js && !aix

package find

import (
	"os"
	"path/filepath"
)

func firefoxRoots(yield func(string, error) bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		_ = yield(``, err)
		return
	}
	// Ubuntu 21.10 (snap)
	if !yield(filepath.Join(home, `snap`, `firefox`, `common`, `.mozilla`, `firefox`), nil) {
		return
	}
	if !yield(filepath.Join(home, `.mozilla`, `firefox`), nil) {
		return
	}
}
