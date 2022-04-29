//+build !windows,!darwin

package opera

import (
	"os"
	"path/filepath"
)

func operaPrestoRoots() ([]string, error) {
	// https://kb.digital-detective.net/display/BF/Location+of+Opera+Presto+Data

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return []string{filepath.Join(home, `.opera`)}, nil
}

func operaBlinkRoots() ([]string, error) {
	// https://kb.digital-detective.net/display/BF/Location+of+Opera+Blink+Data

	var dotConfigs, ret []string

	// fallback
	if home, err := os.UserHomeDir(); err == nil {
		dotConfigs = append(dotConfigs, filepath.Join(home, `.config`))
	}
	if dir, ok := os.LookupEnv(`XDG_CONFIG_HOME`); ok {
		dotConfigs = append(dotConfigs, dir)
	}
	for _, dotConfig := range dotConfigs {
		ret = append(
			ret,
			filepath.Join(dotConfig, `opera`),
		)
	}
	return ret, nil
}
