//go:build plan9 || android || ios || js || aix

package find

import "errors"

func firefoxRoots() ([]string, error) {
	return nil, errors.New(`not implemented`)
}
