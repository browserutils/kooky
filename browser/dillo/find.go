//go:build linux || freebsd || openbsd || netbsd || dragonfly || solaris

package dillo

import (
	"os"
	"path/filepath"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/netscape"
)

type dilloFinder struct{}

var _ kooky.CookieStoreFinder = (*dilloFinder)(nil)

func init() {
	kooky.RegisterFinder(`dillo`, &dilloFinder{})
}

func (f *dilloFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		// https://www.dillo.org/FAQ.html#q16
		// https://www.dillo.org/Cookies.txt

		home, err := os.UserHomeDir()
		if err != nil {
			_ = yield(nil, err)
			return
		}

		st := &cookies.CookieJar{
			CookieStore: &netscape.CookieStore{
				DefaultCookieStore: cookies.DefaultCookieStore{
					BrowserStr:           `dillo`,
					IsDefaultProfileBool: true,
					FileNameStr:          filepath.Join(home, `.dillo`, `cookies.txt`),
				},
			},
		}
		if !yield(st, nil) {
			return
		}
	}
}
