//+build !windows,!darwin,!plan9,!android,!js,!aix

package find

import (
	"os"
	"path/filepath"
)

func firefoxRoots() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return []string{filepath.Join(home, `.mozilla`, `firefox`)}, nil
}
