package opera

import (
	"path/filepath"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/chrome"
	"github.com/browserutils/kooky/internal/cookies"
)

type operaFinder struct{}

var _ kooky.CookieStoreFinder = (*operaFinder)(nil)

func init() {
	kooky.RegisterFinder(`opera`, &operaFinder{})
}

func (f *operaFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		for root, err := range operaPrestoRoots {
			if err != nil {
				if !yield(nil, err) {
					return
				}
			}
			st := &cookies.CookieJar{
				CookieStore: &operaCookieStore{
					CookieStore: &operaPrestoCookieStore{
						DefaultCookieStore: cookies.DefaultCookieStore{
							BrowserStr:           `opera`,
							IsDefaultProfileBool: true,
							FileNameStr:          filepath.Join(root, `cookies4.dat`),
						},
					},
				},
			}
			if !yield(st, nil) {
				return
			}
		}

		for root, err := range operaBlinkRoots {
			if err != nil {
				if !yield(nil, err) {
					return
				}
			}
			st := &cookies.CookieJar{
				CookieStore: &operaCookieStore{
					CookieStore: &chrome.CookieStore{
						DefaultCookieStore: cookies.DefaultCookieStore{
							BrowserStr:           `opera`,
							IsDefaultProfileBool: true,
							FileNameStr:          filepath.Join(root, `Cookies`),
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
