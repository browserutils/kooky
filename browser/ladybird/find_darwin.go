//go:build darwin && !ios

package ladybird

import (
	"os"
	"path/filepath"
)

func ladybirdCookiePaths() []string {
	var paths []string

	if dataDir, ok := os.LookupEnv(`XDG_DATA_HOME`); ok && dataDir != `` {
		paths = append(paths, filepath.Join(dataDir, `Ladybird`, `Ladybird.db`))
	}

	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, `Library`, `Application Support`, `Ladybird`, `Ladybird.db`))
	}

	return paths
}
