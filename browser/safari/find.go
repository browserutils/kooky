//go:build (darwin && !ios) || windows

package safari

import (
	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
)

type safariFinder struct{}

var _ kooky.CookieStoreFinder = (*safariFinder)(nil)

func init() {
	kooky.RegisterFinder(`safari`, &safariFinder{})
}

func (f *safariFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	fileStrs, err := cookieFiles()
	if err != nil {
		return nil, err
	}

	var ret []kooky.CookieStore
	for i, fileStr := range fileStrs {
		ret = append(
			ret,
			&cookies.CookieJar{
				CookieStore: &safariCookieStore{
					DefaultCookieStore: cookies.DefaultCookieStore{
						BrowserStr:           `safari`,
						IsDefaultProfileBool: i == 0,
						FileNameStr:          fileStr,
					},
				},
			},
		)
	}

	return ret, nil
}
