//go:build linux || freebsd || openbsd || netbsd || dragonfly || solaris

package dillo

import (
	"os"
	"path/filepath"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/cookies"
	"github.com/zellyn/kooky/internal/netscape"
)

type dilloFinder struct{}

var _ kooky.CookieStoreFinder = (*dilloFinder)(nil)

func init() {
	kooky.RegisterFinder(`dillo`, &dilloFinder{})
}

func (f *dilloFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	// https://www.dillo.org/FAQ.html#q16
	// https://www.dillo.org/Cookies.txt

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var ret = []kooky.CookieStore{
		&cookies.CookieJar{
			CookieStore: &netscape.CookieStore{
				DefaultCookieStore: cookies.DefaultCookieStore{
					BrowserStr:           `dillo`,
					IsDefaultProfileBool: true,
					FileNameStr:          filepath.Join(home, `.dillo`, `cookies.txt`),
				},
			},
		}}

	return ret, nil
}
