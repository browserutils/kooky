//go:build !windows && !darwin && !plan9 && !android && !ios && !js && !aix

package find

import (
	"os"
	"path/filepath"

	"github.com/browserutils/kooky/internal/wsl"
)

func chromeRoots(yield func(string, error) bool) {
	// "${CHROME_VERSION_EXTRA:-${XDG_CONFIG_HOME:-$HOME/.config}}"
	// https://chromium.googlesource.com/chromium/src.git/+/62.0.3202.58/docs/user_data_dir.md#linux
	var dotConfigs, ret []string
	// fallback
	if home, err := os.UserHomeDir(); err != nil {
		if !yield(``, err) {
			return
		}
	} else {
		dotConfigs = append(dotConfigs, filepath.Join(home, `.config`))
	}
	for _, v := range []string{`XDG_CONFIG_HOME`, `CHROME_CONFIG_HOME`} {
		if dir, ok := os.LookupEnv(v); ok {
			dotConfigs = append(dotConfigs, dir)
		}
	}
	cve, cveOK := os.LookupEnv(`CHROME_VERSION_EXTRA`)
	for _, dotConfig := range dotConfigs {
		ret = append(
			ret,
			filepath.Join(dotConfig, `google-chrome`),
			filepath.Join(dotConfig, `google-chrome-beta`),
			filepath.Join(dotConfig, `google-chrome-unstable`),
		)
		if cveOK {
			ret = append(
				ret,
				filepath.Join(dotConfig, `google-chrome-`+cve),
			)
		}
	}
	for _, r := range ret {
		if !yield(r, nil) {
			return
		}
	}
	// on WSL Linux add Windows paths
	appDataRoot, err := wsl.WSLAppDataRoot()
	if err != nil && !yield(``, err) {
		return
	}
	for r, err := range windowsChromeRoots(filepath.Join(appDataRoot, `Local`)) {
		if !yield(r, err) {
			return
		}
	}
}

func chromiumRoots(yield func(string, error) bool) {
	// "${XDG_CONFIG_HOME:-$HOME/.config}"
	var dotConfigs []string
	// fallback
	if home, err := os.UserHomeDir(); err != nil {
		if !yield(``, err) {
			return
		}
	} else {
		dotConfigs = append(dotConfigs, filepath.Join(home, `.config`))
	}
	if dir, ok := os.LookupEnv(`XDG_CONFIG_HOME`); ok {
		dotConfigs = append(dotConfigs, dir)
	}
	for _, dotConfig := range dotConfigs {
		if !yield(filepath.Join(dotConfig, `chromium`), nil) {
			return
		}
	}
	// on WSL Linux add Windows paths
	appDataRoot, err := wsl.WSLAppDataRoot()
	if err != nil && !yield(``, err) {
		return
	}
	for r, err := range windowsChromiumRoots(filepath.Join(appDataRoot, `Local`)) {
		if !yield(r, err) {
			return
		}
	}
}
