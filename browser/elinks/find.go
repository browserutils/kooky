package elinks

import (
	"os"
	"path/filepath"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal"
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

	var s elinksCookieStore
	d := internal.DefaultCookieStore{
		BrowserStr:           `elinks`,
		IsDefaultProfileBool: true,
		FileNameStr:          filepath.Join(home, `.elinks`, `cookies`),
	}
	internal.SetCookieStore(&d, &s)
	s.DefaultCookieStore = d
	var ret = []kooky.CookieStore{&s}

	return ret, nil
}
