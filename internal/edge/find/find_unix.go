//go:build !windows && !darwin && !plan9 && !android && !js && !aix

package find

import (
	"os"
	"path/filepath"
)

func edgeRoots() ([]string, error) {
	return nil, errors.New(`not implemented`)
}
