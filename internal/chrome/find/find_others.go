//go:build plan9 || ios || js || aix

package find

import "errors"

var errNotImplemented = errors.New(`not implemented`)

func chromeRoots() ([]string, error) {
	return nil, errNotImplemented
}

func chromiumRoots() ([]string, error) {
	return nil, errNotImplemented
}
