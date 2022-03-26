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

func (f *dilloFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	// https://www.dillo.org/FAQ.html#q16
	// https://www.dillo.org/Cookies.txt

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var s netscape.CookieStore
	d := internal.DefaultCookieStore{
		BrowserStr:           `dillo`,
		IsDefaultProfileBool: true,
		FileNameStr:          filepath.Join(home, `.dillo`, `cookies.txt`),
	}
	internal.SetCookieStore(&d, &s)
	s.DefaultCookieStore = d
	var ret = []kooky.CookieStore{&s}

	return ret, nil
}
