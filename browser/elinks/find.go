//go:build linux || freebsd || openbsd || netbsd || dragonfly || solaris

package elinks

import (
	"os"
	"path/filepath"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/cookies"
)

type elinksFinder struct{}

var _ kooky.CookieStoreFinder = (*elinksFinder)(nil)

func init() {
	kooky.RegisterFinder(`elinks`, &elinksFinder{})
}

func (f *elinksFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var ret = []kooky.CookieStore{
		&cookies.CookieJar{
			CookieStore: &elinksCookieStore{
				DefaultCookieStore: cookies.DefaultCookieStore{
					BrowserStr:           `elinks`,
					IsDefaultProfileBool: true,
					FileNameStr:          filepath.Join(home, `.elinks`, `cookies`),
				},
			},
		},
	}

	return ret, nil
}
