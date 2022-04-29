//go:build darwin && !ios

package safari

import (
	"os"
	"path/filepath"
)

func cookieFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return ``, err
	}
	return filepath.Join(home, `Library`, `Cookies`, `Cookies.binarycookies`), nil
}
