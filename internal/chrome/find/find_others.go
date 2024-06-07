//go:build plan9 || ios || js || aix

package find

import "errors"

var errNotImplemented = errors.New(`not implemented`)

func chromeRoots(yield func(string, error) bool) { _ = yield(``, errNotImplemented) }

func chromiumRoots(yield func(string, error) bool) { _ = yield(``, errNotImplemented) }
