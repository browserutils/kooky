package ie

import (
	"github.com/xiazemin/kooky"
	"github.com/xiazemin/kooky/internal/cookies"
	"github.com/xiazemin/kooky/internal/ie"
	_ "github.com/xiazemin/kooky/internal/ie/find"
)

type ieFinder struct{}

var _ kooky.CookieStoreFinder = (*ieFinder)(nil)

func init() { kooky.RegisterFinder(`ie`, &ieFinder{}) }

func (f *ieFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		for root, err := range ieRoots {
			if err != nil {
				_ = yield(nil, err)
				return
			}
			st := &cookies.CookieJar{
				CookieStore: &ie.CookieStore{
					CookieStore: &ie.TextCookieStore{
						DefaultCookieStore: cookies.DefaultCookieStore{
							BrowserStr:           `ie`,
							IsDefaultProfileBool: true,
							FileNameStr:          root,
						},
					},
				},
			}
			if !yield(st, nil) {
				return
			}
		}
	}
}
