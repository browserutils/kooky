package dillo

import (
	"os"
	"path/filepath"

	"github.com/zellyn/kooky"
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
		&dilloCookieStore{
			browser:          `dillo`,
			isDefaultProfile: true,
			filename:         filepath.Join(home, `.dillo`, `cookies.txt`),
		},
	}
	return ret, nil
}
