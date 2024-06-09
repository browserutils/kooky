//go:build linux || freebsd || openbsd || netbsd || dragonfly || solaris

package browsh

import (
	"os"
	"path/filepath"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/firefox"
)

type browshFinder struct{}

var _ kooky.CookieStoreFinder = (*browshFinder)(nil)

func init() {
	kooky.RegisterFinder(`browsh`, &browshFinder{})
}

func (f *browshFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		dotConfig, err := os.UserConfigDir()
		if err != nil {
			_ = yield(nil, err)
			return
		}

		st := &cookies.CookieJar{
			CookieStore: &firefox.CookieStore{
				DefaultCookieStore: cookies.DefaultCookieStore{
					BrowserStr:           `browsh`,
					IsDefaultProfileBool: true,
					FileNameStr:          filepath.Join(dotConfig, `browsh`, `firefox_profile`, `cookies.sqlite`),
				},
			},
		}
		if !yield(st, nil) {
			return
		}
	}
}
