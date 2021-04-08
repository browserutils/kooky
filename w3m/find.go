package w3m

import (
	"os"
	"path/filepath"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal"
)

type w3mFinder struct{}

var _ kooky.CookieStoreFinder = (*w3mFinder)(nil)

func init() {
	kooky.RegisterFinder(`w3m`, &w3mFinder{})
}

func (s *w3mFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var ret = []kooky.CookieStore{
		&w3mCookieStore{
			DefaultCookieStore: internal.DefaultCookieStore{
				BrowserStr:           `w3m`,
				IsDefaultProfileBool: true,
				FileNameStr:          filepath.Join(home, `.w3m`, `cookie`),
			},
		},
	}
	return ret, nil
}
