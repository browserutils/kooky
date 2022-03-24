//go:build windows

// Safari v5.1.7 was the last version for Windows

package safari

import (
	"os"
	"path/filepath"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal"
)

type safariFinder struct{}

var _ kooky.CookieStoreFinder = (*safariFinder)(nil)

func init() {
	kooky.RegisterFinder(`safari`, &safariFinder{})
}

func (s *safariFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	confDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	var ret = []kooky.CookieStore{
		&safariCookieStore{
			DefaultCookieStore: internal.DefaultCookieStore{
				BrowserStr:           `safari`,
				IsDefaultProfileBool: true,
				FileNameStr:          filepath.Join(confDir, `Apple Computer`, `Safari`, `Cookies`, `Cookies.binarycookies`),
			},
		},
	}
	return ret, nil
}
