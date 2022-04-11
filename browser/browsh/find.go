//go:build linux || freebsd || openbsd || netbsd || dragonfly || solaris

package browsh

import (
	"os"
	"path/filepath"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/cookies"
	"github.com/zellyn/kooky/internal/firefox"
)

type browshFinder struct{}

var _ kooky.CookieStoreFinder = (*browshFinder)(nil)

func init() {
	kooky.RegisterFinder(`browsh`, &browshFinder{})
}

func (f *browshFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	dotConfig, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	var ret = []kooky.CookieStore{
		&cookies.CookieJar{
			CookieStore: &firefox.CookieStore{
				DefaultCookieStore: cookies.DefaultCookieStore{
					BrowserStr:           `browsh`,
					IsDefaultProfileBool: true,
					FileNameStr:          filepath.Join(dotConfig, `browsh`, `firefox_profile`, `cookies.sqlite`),
				},
			},
		},
	}

	return ret, nil
}
