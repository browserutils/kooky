//go:build windows
// +build windows

package opera

import (
	"errors"
	"os"
	"path/filepath"
)

func operaPrestoRoots(yield func(string, error) bool) {
	// https://kb.digital-detective.net/display/BF/Location+of+Opera+Presto+Data
	appData, ok := os.LookupEnv(`AppData`)
	if !ok {
		_ = yield(``, errors.New(`%AppData% not set`))
		return
	}
	pathEnds := [][]string{
		{`Opera`, `Opera`}, // TODO check
	}
	for _, end := range pathEnds {
		if !yield(filepath.Join(append([]string{appData}, end...)...), nil) {
			return
		}
	}
}

func operaBlinkRoots(yield func(string, error) bool) {
	// Windows XP: %HOMEPATH%\Application Data\Opera Software\Opera Stable\
	// Windows 7, 8: %AppData%\Opera Software\Opera Stable\
	// https://kb.digital-detective.net/display/BF/Location+of+Opera+Blink+Data

	appData, ok := os.LookupEnv(`AppData`)
	if !ok {
		_ = yield(``, errors.New(`%AppData% not set`))
		return
	}
	pathEnds := [][]string{
		{`Opera Software`, `Opera Stable`},
	}
	for _, end := range pathEnds {
		if !yield(filepath.Join(append([]string{appData}, end...)...), nil) {
			return
		}
	}
}
