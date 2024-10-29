//go:build windows || darwin || plan9 || android || js || aix
// +build windows darwin plan9 android js aix

package netscape

import "errors"

func netscapeRoots(yield func(string, error) bool) { yield(``, errors.New(`not implemented`)) }
