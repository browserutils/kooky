//go:build !windows && !linux

package find

func windowsFirefoxRoots(yield func(string, error) bool) {}
