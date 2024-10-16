package chrome

import (
	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/chrome"
	"github.com/browserutils/kooky/internal/chrome/find"
	"github.com/browserutils/kooky/internal/cookies"
)

type chromeFinder struct{}

var _ kooky.CookieStoreFinder = (*chromeFinder)(nil)

func init() {
	kooky.RegisterFinder(`chrome`, &chromeFinder{})
}

func (f *chromeFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		for file, err := range find.FindChromeCookieStoreFiles() {
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
