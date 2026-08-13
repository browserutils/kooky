//go:build !windows && !linux

package find

func windowsZenRoots(yield func(string, error) bool) {}
