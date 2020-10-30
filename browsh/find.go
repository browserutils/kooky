package browsh

import (
	"os"
	"path/filepath"

	"github.com/zellyn/kooky"
)

type browshFinder struct{}

var _ kooky.CookieStoreFinder = (*browshFinder)(nil)

func init() {
	kooky.RegisterFinder(`browsh`, &browshFinder{})
}

func (s *browshFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	dotConfig, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	var ret = []kooky.CookieStore{
		&browshCookieStore{
			browser:          `browsh`,
			isDefaultProfile: true,
			filename:         filepath.Join(dotConfig, `browsh`, `firefox_profile`, `cookies.sqlite`),
		},
	}

	return ret, nil
}
