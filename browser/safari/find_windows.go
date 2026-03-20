//go:build windows

// Safari v5.1.7 was the last version for Windows

package safari

import (
	"path/filepath"

	"github.com/browserutils/kooky/internal/windowsx"
)

func cookieFiles() ([]string, error) {
	appData, err := windowsx.AppData()
	if err != nil {
		return nil, err
	}
	return []string{filepath.Join(appData, `Apple Computer`, `Safari`, `Cookies`, `Cookies.binarycookies`)}, nil
}
