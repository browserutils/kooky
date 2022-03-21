//+build windows

package opera

import (
	"errors"
	"os"
	"path/filepath"
)

func operaPrestoRoots() ([]string, error) {
	// https://kb.digital-detective.net/display/BF/Location+of+Opera+Presto+Data
	appData, ok := os.LookupEnv(`AppData`)
	if !ok {
		return nil, errors.New(`%AppData% not set`)
	}
	var ret []string
	pathEnds := [][]string{
		{`Opera`, `Opera`},
	}
	for _, end := range pathEnds {
		ret = append(
			ret,
			filepath.Join(append([]string{appData}, end...)...),
		)
	}
	return ret, nil
}

func operaBlinkRoots() ([]string, error) {
	// Windows XP: %HOMEPATH%\Application Data\Opera Software\Opera Stable\
	// Windows 7, 8: %AppData%\Opera Software\Opera Stable\
	// https://kb.digital-detective.net/display/BF/Location+of+Opera+Blink+Data

	appData, ok := os.LookupEnv(`AppData`)
	if !ok {
		return nil, errors.New(`%AppData% not set`)
	}
	var ret []string
	pathEnds := [][]string{
		{`Opera Software`, `Opera Stable`},
	}
	for _, end := range pathEnds {
		ret = append(
			ret,
			filepath.Join(append([]string{appData}, end...)...),
		)
	}
	return ret, nil
}
