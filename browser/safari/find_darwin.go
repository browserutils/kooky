//go:build darwin && !ios

package safari

import (
	"os"
	"path/filepath"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func cookieFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return ``, err
	}
	paths := []string{
		filepath.Join(home, `Library`, `Cookies`, `Cookies.binarycookies`),
		// ~/Library/Containers/com.apple.Safari/Data/Library/Cookies
		filepath.Join(home, `Library`, `Containers`, `com.apple.Safari`, `Data`, `Library`, `Cookies`, `Cookies.binarycookies`),
	}
	for _, path := range paths {
		if fileExists(path) {
			return path, nil
		}
	}

	return ``, os.ErrNotExist
}
