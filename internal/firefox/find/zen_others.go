//go:build plan9 || android || ios || js || aix

package find

import "errors"

func zenRoots(yield func(string, error) bool) { yield(``, errors.New(`not implemented`)) }
