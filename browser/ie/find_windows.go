//go:build windows

package ie

import (
	"os"
	"path/filepath"
)

func ieRoots() ([]string, error) {
	confDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	return []string{
		filepath.Join(confDir, `Microsoft`, `Windows`, `Cookies`),
		filepath.Join(confDir, `Microsoft`, `Windows`, `Cookies`, `Low`),
	}, nil
}
