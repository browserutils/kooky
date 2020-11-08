package netscape

import (
	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal"
	"github.com/zellyn/kooky/internal/firefox/find"
	"github.com/zellyn/kooky/internal/netscape"
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
			&netscape.CookieStore{
				DefaultCookieStore: internal.DefaultCookieStore{
					BrowserStr:           file.Browser,
					ProfileStr:           file.Profile,
					IsDefaultProfileBool: file.IsDefaultProfile,
					FileNameStr:          file.Path,
				},
			},
		)
	}

	return ret, nil
}
