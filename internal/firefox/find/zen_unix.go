//go:build !windows && !darwin && !plan9 && !android && !js && !aix

package find

import (
	"os"
	"path/filepath"

	"github.com/browserutils/kooky/internal/windowsx"
)

func zenRoots(yield func(string, error) bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		_ = yield(``, err)
		return
	}
	// tarball / AUR install, current default:
	// https://docs.zen-browser.app/guides/install-linux
	if !yield(filepath.Join(home, `.zen`), nil) {
		return
	}
	// Flatpak (same doc)
	if !yield(filepath.Join(home, `.var`, `app`, `app.zen_browser.zen`, `.zen`), nil) {
		return
	}
	// Newer XDG Base Directory-following builds
	// (zen-browser/desktop#12366) — gated behind the Twilight channel as of
	// this writing, kept here so this doesn't need a follow-up once it
	// reaches the stable channel.
	if cfgDir, err := os.UserConfigDir(); err == nil {
		if !yield(filepath.Join(cfgDir, `zen`), nil) {
			return
		}
	}
	// on WSL Linux add Windows paths
	if !windowsx.IsWSL() {
		return
	}
	for r, err := range windowsZenRoots {
		if !yield(r, err) {
			return
		}
	}
}
