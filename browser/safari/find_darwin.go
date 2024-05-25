//go:build darwin && !ios

package safari

import (
	"os"
	"path/filepath"
)

func cookieFiles() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	paths := []string{
		// ~/Library/Containers/com.apple.Safari/Data/Library/Cookies
		filepath.Join(home, `Library`, `Containers`, `com.apple.Safari`, `Data`, `Library`, `Cookies`, `Cookies.binarycookies`),
		filepath.Join(home, `Library`, `Cookies`, `Cookies.binarycookies`),
	}

	return paths, nil
}
