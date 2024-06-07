//go:build darwin
// +build darwin

package opera

import (
	"os"
	"path/filepath"
)

func operaPrestoRoots(yield func(string, error) bool) {
	// https://kb.digital-detective.net/display/BF/Location+of+Opera+Presto+Data

	home, err := os.UserHomeDir()
	if err != nil {
		_ = yield(``, err)
		return
	}

	// /Users/{user}/Library/Opera/
	if !yield(filepath.Join(home, `Library`, `Opera`), nil) {
		return
	}
}

func operaBlinkRoots(yield func(string, error) bool) {
	// https://kb.digital-detective.net/display/BF/Location+of+Opera+Blink+Data

	home, err := os.UserHomeDir()
	if err != nil {
		_ = yield(``, err)
		return
	}

	// /Users/{user}/Library/Application Support/com.operasoftware.Opera/
	if !yield(filepath.Join(home, `Library`, `Application Support`, `com.operasoftware.Opera`), nil) {
		return
	}
}
