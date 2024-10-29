//go:build !windows && !darwin && !plan9 && !android && !js && !aix
// +build !windows,!darwin,!plan9,!android,!js,!aix

package netscape

import (
	"os"
	"path/filepath"
)

func netscapeRoots(yield func(string, error) bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		_ = yield(``, err)
		return
	}
	if !yield(filepath.Join(home, `.netscape`, `navigator`), nil) {
		return
	}
}
