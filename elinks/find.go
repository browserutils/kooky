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

func (s *elinksFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var ret = []kooky.CookieStore{
		&elinksCookieStore{
			DefaultCookieStore: internal.DefaultCookieStore{
				BrowserStr:           `elinks`,
				IsDefaultProfileBool: true,
				FileNameStr:          filepath.Join(home, `.elinks`, `cookies`),
			},
		},
	}
	return ret, nil
}
