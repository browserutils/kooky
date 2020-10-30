//+build darwin

package opera

import (
	"os"
	"path/filepath"
)

func operaPrestoRoots() ([]string, error) {
	// https://kb.digital-detective.net/display/BF/Location+of+Opera+Presto+Data

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// /Users/{user}/Library/Opera/
	return []string{filepath.Join(home, `Library`, `Opera`)}, nil
}

func operaBlinkRoots() ([]string, error) {
	// https://kb.digital-detective.net/display/BF/Location+of+Opera+Blink+Data

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// /Users/{user}/Library/Application Support/com.operasoftware.Opera/
	return []string{filepath.Join(home, `Library`, `Application Support`, `com.operasoftware.Opera`)}, nil
}
