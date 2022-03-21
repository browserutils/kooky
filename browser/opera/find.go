package opera

import (
	"path/filepath"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal"
	"github.com/zellyn/kooky/internal/chrome"
)

type operaFinder struct{}

var _ kooky.CookieStoreFinder = (*operaFinder)(nil)

func init() {
	kooky.RegisterFinder(`opera`, &operaFinder{})
}

func (s *operaFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	var files []kooky.CookieStore

	roots, err := operaPrestoRoots()
	if err != nil {
		return nil, err
	}
	for _, root := range roots {
		files = append(
			files,
			&operaCookieStore{
				CookieStore: chrome.CookieStore{
					DefaultCookieStore: internal.DefaultCookieStore{
						BrowserStr:           `opera`,
						IsDefaultProfileBool: true,
						FileNameStr:          filepath.Join(root, `cookies4.dat`),
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
		files = append(
			files,
			&operaCookieStore{
				CookieStore: chrome.CookieStore{
					DefaultCookieStore: internal.DefaultCookieStore{
						BrowserStr:           `opera`,
						IsDefaultProfileBool: true,
						FileNameStr:          filepath.Join(root, `Cookies`),
					},
				},
			},
		)
	}

	return files, nil
}
