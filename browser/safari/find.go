//go:build (darwin && !ios) || windows

package safari

import (
	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/cookies"
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

	var ret = []kooky.CookieStore{
		&cookies.CookieJar{
			CookieStore: &safariCookieStore{
				DefaultCookieStore: cookies.DefaultCookieStore{
					BrowserStr:           `safari`,
					IsDefaultProfileBool: true,
					FileNameStr:          fileStr,
				},
			},
		},
	}

	return ret, nil
}
