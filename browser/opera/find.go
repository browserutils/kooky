package opera

import (
	"path/filepath"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/chrome"
	"github.com/zellyn/kooky/internal/cookies"
)

type operaFinder struct{}

var _ kooky.CookieStoreFinder = (*operaFinder)(nil)

func init() {
	kooky.RegisterFinder(`opera`, &operaFinder{})
}

func (f *operaFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	var ret []kooky.CookieStore

	roots, err := operaPrestoRoots()
	if err != nil {
		return nil, err
	}
	for _, root := range roots {
		ret = append(
			ret,
			&cookies.CookieJar{
				CookieStore: &operaCookieStore{
					CookieStore: &operaPrestoCookieStore{
						DefaultCookieStore: cookies.DefaultCookieStore{
							BrowserStr:           `opera`,
							IsDefaultProfileBool: true,
							FileNameStr:          filepath.Join(root, `cookies4.dat`),
						},
					},
				},
			},
		)
	}

	roots, err = operaBlinkRoots()
	if err != nil {
		return nil, err
	}
	for _, root := range roots {
		ret = append(
			ret,
			&cookies.CookieJar{
				CookieStore: &operaCookieStore{
					CookieStore: &chrome.CookieStore{
						DefaultCookieStore: cookies.DefaultCookieStore{
							BrowserStr:           `opera`,
							IsDefaultProfileBool: true,
							FileNameStr:          filepath.Join(root, `Cookies`),
						},
					},
				},
			},
		)
	}

	return ret, nil
}
