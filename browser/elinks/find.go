//go:build linux || freebsd || openbsd || netbsd || dragonfly || solaris

package elinks

import (
	"os"
	"path/filepath"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
)

type elinksFinder struct{}

var _ kooky.CookieStoreFinder = (*elinksFinder)(nil)

func init() {
	kooky.RegisterFinder(`elinks`, &elinksFinder{})
}

func (f *elinksFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		home, err := os.UserHomeDir()
		if err != nil {
			_ = yield(nil, err)
			return
		}

		st := &cookies.CookieJar{
			CookieStore: &elinksCookieStore{
				DefaultCookieStore: cookies.DefaultCookieStore{
					BrowserStr:           `elinks`,
					IsDefaultProfileBool: true,
					FileNameStr:          filepath.Join(home, `.elinks`, `cookies`),
				},
			},
		}
		if !yield(st, nil) {
			return
		}
	}
}
