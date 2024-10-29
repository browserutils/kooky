//go:build windows

// Safari v5.1.7 was the last version for Windows

package safari

import (
	"os"
	"path/filepath"
)

func cookieFiles() ([]string, error) {
	confDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return []string{filepath.Join(confDir, `Apple Computer`, `Safari`, `Cookies`, `Cookies.binarycookies`)}, nil
}
