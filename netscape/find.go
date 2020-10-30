package netscape

import (
	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/firefox/find"
)

type netscapeFinder struct{}

var _ kooky.CookieStoreFinder = (*netscapeFinder)(nil)

func init() {
	kooky.RegisterFinder(`netscape`, &netscapeFinder{})
}

func (s *netscapeFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	files, err := find.FindCookieStoreFiles(netscapeRoots, `netscape`, `cookies.txt`)
	if err != nil {
		return nil, err
	}

	var ret []kooky.CookieStore
	for _, file := range files {
		ret = append(
			ret,
			&netscapeCookieStore{
				filename:         file.Path,
				browser:          file.Browser,
				profile:          file.Profile,
				isDefaultProfile: file.IsDefaultProfile,
			},
		)
	}

	return ret, nil
}
