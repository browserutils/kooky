//go:build plan9 || android || ios || js || aix

package find

import "errors"

var errNotImplemented = errors.New(`not implemented`)

func chromeRoots() ([]string, error) {
	return nil, errNotImplemented
}

func chromiumRoots() ([]string, error) {
	return nil, errNotImplemented
}
