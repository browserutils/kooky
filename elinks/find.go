package elinks

import (
	"os"
	"path/filepath"

	"github.com/zellyn/kooky"
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
			browser:          `elinks`,
			isDefaultProfile: true,
			filename:         filepath.Join(home, `.elinks`, `cookies`),
		},
	}
	return ret, nil
}
