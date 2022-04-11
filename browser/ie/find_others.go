//go:build !windows

package ie

import "errors"

// TODO

func ieRoots() ([]string, error) {
	return nil, errors.New(`not implemented`)
}
