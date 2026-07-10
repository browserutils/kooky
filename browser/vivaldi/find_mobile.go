//go:build ios || android

package vivaldi

import "errors"

func vivaldiRoots(yield func(string, error) bool) { yield(``, errors.New(`not implemented`)) }
