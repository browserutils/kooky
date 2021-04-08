package dillo

import (
	"os"
	"path/filepath"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal"
	"github.com/zellyn/kooky/internal/netscape"
)

type dilloFinder struct{}

var _ kooky.CookieStoreFinder = (*dilloFinder)(nil)

func init() {
	kooky.RegisterFinder(`dillo`, &dilloFinder{})
}

func (s *dilloFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	// https://www.dillo.org/FAQ.html#q16
	// https://www.dillo.org/Cookies.txt

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var ret = []kooky.CookieStore{
		&netscape.CookieStore{
			DefaultCookieStore: internal.DefaultCookieStore{
				BrowserStr:           `dillo`,
				IsDefaultProfileBool: true,
				FileNameStr:          filepath.Join(home, `.dillo`, `cookies.txt`),
			},
		},
	}
	return ret, nil
}
