//go:build (darwin && !ios) || windows

package safari

import (
	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal"
)

type safariFinder struct{}

var _ kooky.CookieStoreFinder = (*safariFinder)(nil)

func init() {
	kooky.RegisterFinder(`safari`, &safariFinder{})
}

func (f *safariFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	fileStr, err := cookieFile()
	if err != nil {
		return nil, err
	}

	var s safariCookieStore
	d := internal.DefaultCookieStore{
		BrowserStr:           `safari`,
		IsDefaultProfileBool: true,
		FileNameStr:          fileStr,
	}
	internal.SetCookieStore(&d, &s)
	s.DefaultCookieStore = d

	var ret = []kooky.CookieStore{&s}
	return ret, nil
}
