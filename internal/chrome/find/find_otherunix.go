//go:build !linux && !windows && !darwin && !plan9 && !android && !ios && !js && !aix

package find

func windowsChromeRoots(_ string) func(yield func(string, error) bool) {
	return func(yield func(string, error) bool) {}
}

func windowsChromiumRoots(_ string) func(yield func(string, error) bool) {
	return func(yield func(string, error) bool) {}
}
