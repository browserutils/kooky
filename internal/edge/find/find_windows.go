//go:build windows
// +build windows

package find

import (
	"errors"
	"os"
	"path/filepath"
)

func edgeRoots() ([]string, error) {
	// AppData Local
	locApp := os.Getenv(`LocalAppData`)
	if len(locApp) == 0 {
		return nil, errors.New(`%LocalAppData% is empty`)
	}

	var ret = []string{
		filepath.Join(locApp, `Microsoft`, `Edge`, `User Data`),
	}

	return ret, nil
}
