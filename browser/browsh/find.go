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

func (s *browshFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	dotConfig, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	var ret = []kooky.CookieStore{
		&firefox.CookieStore{
			DefaultCookieStore: internal.DefaultCookieStore{
				BrowserStr:           `browsh`,
				IsDefaultProfileBool: true,
				FileNameStr:          filepath.Join(dotConfig, `browsh`, `firefox_profile`, `cookies.sqlite`),
			},
		},
	}

	return ret, nil
}
