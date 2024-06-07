//go:build !windows && !darwin
// +build !windows,!darwin

package opera

import (
	"os"
	"path/filepath"
)

func operaPrestoRoots(yield func(string, error) bool) {
	// https://kb.digital-detective.net/display/BF/Location+of+Opera+Presto+Data

	home, err := os.UserHomeDir()
	if err != nil {
		_ = yield(``, err)
		return
	}

	_ = yield(filepath.Join(home, `.opera`), nil)
}

func operaBlinkRoots(yield func(string, error) bool) {
	// https://kb.digital-detective.net/display/BF/Location+of+Opera+Blink+Data

	var dotConfigs []string

	// fallback
	if home, err := os.UserHomeDir(); err == nil {
		dotConfigs = append(dotConfigs, filepath.Join(home, `.config`))
	}
	if dir, ok := os.LookupEnv(`XDG_CONFIG_HOME`); ok {
		dotConfigs = append(dotConfigs, dir)
	}
	for _, dotConfig := range dotConfigs {
		if !yield(filepath.Join(dotConfig, `opera`), nil) {
			return
		}
	}
}
