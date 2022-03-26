package opera

import (
	"path/filepath"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal"
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
		var s operaCookieStore
		d := internal.DefaultCookieStore{
			BrowserStr:           `opera`,
			IsDefaultProfileBool: true,
			FileNameStr:          filepath.Join(root, `cookies4.dat`),
		}
		internal.SetCookieStore(&d, &s)
		s.DefaultCookieStore = d
		ret = append(ret, &s)
	}

	roots, err = operaBlinkRoots()
	if err != nil {
		return nil, err
	}
	for _, root := range roots {
		var s operaCookieStore
		d := internal.DefaultCookieStore{
			BrowserStr:           `opera`,
			IsDefaultProfileBool: true,
			FileNameStr:          filepath.Join(root, `Cookies`),
		}
		internal.SetCookieStore(&d, &s)
		s.DefaultCookieStore = d
		ret = append(ret, &s)
	}

	return ret, nil
}
