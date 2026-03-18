//go:build linux || freebsd || openbsd || netbsd || dragonfly || solaris

package browsh

import (
	"os"
	"path/filepath"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/firefox"
	"github.com/browserutils/kooky/internal/firefox/find"
)

type browshFinder struct{}

var _ kooky.CookieStoreFinder = (*browshFinder)(nil)

func init() {
	kooky.RegisterFinder(`browsh`, &browshFinder{})
}

func (f *browshFinder) FindCookieStores() kooky.CookieStoreSeq {
	dotConfig, err := os.UserConfigDir()
	if err != nil {
		return func(yield func(kooky.CookieStore, error) bool) {
			_ = yield(nil, err)
		}
	}

	profiles := func(yield func(find.Profile, error) bool) {
		_ = yield(find.Profile{
			Path:             filepath.Join(dotConfig, `browsh`, `firefox_profile`),
			Browser:          `browsh`,
			IsDefaultProfile: true,
		}, nil)
	}
	return firefox.CookieStoresForProfiles(profiles)
}
