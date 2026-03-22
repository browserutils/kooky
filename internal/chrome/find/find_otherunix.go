//go:build !linux && !windows && !darwin && !plan9 && !android && !ios && !js && !aix

package find

// for Windows and WSL

func windowsChromeRoots(yield func(string, error) bool)   {}
func windowsChromiumRoots(yield func(string, error) bool) {}
func windowsBraveRoots(yield func(string, error) bool)    {}
