package brave

import (
	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/chrome"
	"github.com/browserutils/kooky/internal/chrome/find"
	"github.com/browserutils/kooky/internal/cookies"
)

type braveFinder struct{}

var _ kooky.CookieStoreFinder = (*braveFinder)(nil)

func init() {
	kooky.RegisterFinder(`brave`, &braveFinder{})
}

func (f *braveFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		for file, err := range find.FindBraveCookieStoreFiles() {
			if err != nil {
				if !yield(nil, err) {
					return
				}
				continue
			}
			st := &cookies.CookieJar{
				CookieStore: &chrome.CookieStore{
					DefaultCookieStore: cookies.DefaultCookieStore{
						BrowserStr:           file.Browser,
						ProfileStr:           file.Profile,
						OSStr:                file.OS,
						IsDefaultProfileBool: file.IsDefaultProfile,
						FileNameStr:          file.Path,
					},
				},
			}
			if !yield(st, nil) {
				return
			}
		}
	}
}
