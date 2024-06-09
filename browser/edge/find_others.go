//go:build !windows && !darwin && !linux

package edge

import "errors"

func edgeChromiumRoots(yield func(string, error) bool) {
	_ = yield(``, errors.New(`platform not supported`))
}
