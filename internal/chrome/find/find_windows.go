//go:build windows
// +build windows

package find

import (
	"os"
)

var (
	chromeRoots   = windowsChromeRoots(os.Getenv(`LocalAppData`))
	chromiumRoots = windowsChromiumRoots(os.Getenv(`LocalAppData`))
)
