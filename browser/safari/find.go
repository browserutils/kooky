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

func (f *safariFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		fileStrs, err := cookieFiles()
		if err != nil {
			_ = yield(nil, err)
			return
		}

		for i, fileStr := range fileStrs {
			st := &cookies.CookieJar{
				CookieStore: &safariCookieStore{
					DefaultCookieStore: cookies.DefaultCookieStore{
						BrowserStr:           `safari`,
						IsDefaultProfileBool: i == 0,
						FileNameStr:          fileStr,
					},
				},
			}
			if !yield(st, nil) {
				return
			}
		}
	}
}
