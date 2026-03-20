//go:build windows || linux

package opera

import (
	"path/filepath"

	"github.com/browserutils/kooky/internal/windowsx"
)

// for Windows and WSL

func windowsOperaPrestoRoots(yield func(string, error) bool) {
	// https://kb.digital-detective.net/display/BF/Location+of+Opera+Presto+Data
	appData, err := windowsx.AppData()
	if err != nil {
		_ = yield(``, err)
		return
	}
	_ = yield(filepath.Join(appData, `Opera`, `Opera`), nil)
}

func windowsOperaBlinkRoots(yield func(string, error) bool) {
	// Windows XP: %HOMEPATH%\Application Data\Opera Software\Opera Stable\
	// Windows 7, 8: %AppData%\Opera Software\Opera Stable\
	// https://kb.digital-detective.net/display/BF/Location+of+Opera+Blink+Data
	appData, err := windowsx.AppData()
	if err != nil {
		_ = yield(``, err)
		return
	}
	_ = yield(filepath.Join(appData, `Opera Software`, `Opera Stable`), nil)
}
