package chromium

import (
	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/chrome"
	"github.com/browserutils/kooky/internal/chrome/find"
	"github.com/browserutils/kooky/internal/cookies"
)

type chromiumFinder struct{}

var _ kooky.CookieStoreFinder = (*chromiumFinder)(nil)

func init() {
	kooky.RegisterFinder(`chromium`, &chromiumFinder{})
}

func (f *chromiumFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		for file, err := range find.FindChromiumCookieStoreFiles() {
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
