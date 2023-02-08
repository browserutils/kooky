//go:build plan9 || android || ios || js || aix

package find

import "errors"

func edgeRoots() ([]string, error) {
	return nil, errors.New(`not implemented`)
}
