//go:build !windows && !darwin && !linux

package vivaldi

import "errors"

func vivaldiRoots(yield func(string, error) bool) {
	_ = yield(``, errors.New(`platform not supported`))
}

func windowsVivaldiRoots(yield func(string, error) bool) {}
