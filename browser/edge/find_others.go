//go:build !windows && !darwin && !linux

package edge

import "errors"

func edgeChromiumRoots() ([]string, error) {
	return nil, errors.New(`platform not supported`)
}
