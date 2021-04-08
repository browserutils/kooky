//+build !windows,!darwin,!plan9,!android,!js,!aix

package netscape

import (
	"os"
	"path/filepath"
)

func netscapeRoots() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return []string{filepath.Join(home, `.netscape`, `navigator`)}, nil
}
