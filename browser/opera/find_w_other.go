//go:build !windows && !linux

package opera

func windowsOperaPrestoRoots(yield func(string, error) bool) {}

func windowsOperaBlinkRoots(yield func(string, error) bool) {}
