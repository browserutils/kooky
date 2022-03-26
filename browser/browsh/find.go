package browsh

import (
	"os"
	"path/filepath"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal"
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

	var s firefox.CookieStore
	d := internal.DefaultCookieStore{
		BrowserStr:           `browsh`,
		IsDefaultProfileBool: true,
		FileNameStr:          filepath.Join(dotConfig, `browsh`, `firefox_profile`, `cookies.sqlite`),
	}
	internal.SetCookieStore(&d, &s)
	s.DefaultCookieStore = d
	var ret = []kooky.CookieStore{&s}

	return ret, nil
}
