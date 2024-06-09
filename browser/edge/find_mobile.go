//go:build ios || android

package edge

import "errors"

func edgeChromiumRoots(yield func(string, error) bool) { yield(``, errors.New(`not implemented`)) }
