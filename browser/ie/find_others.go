//go:build !windows

package ie

import "errors"

// TODO

func ieRoots(yield func(string, error) bool) { _ = yield(``, errors.New(`not implemented`)) }
