//go:build darwin && !ios

package dia

import (
	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/chrome"
	chromefind "github.com/browserutils/kooky/internal/chrome/find"
	"github.com/browserutils/kooky/internal/cookies"
)

type diaFinder struct{}

var _ kooky.CookieStoreFinder = (*diaFinder)(nil)

func init() {
	kooky.RegisterFinder(`dia`, &diaFinder{})
}

func (f *diaFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		for file, err := range chromefind.FindCookieStoreFiles(diaChromiumRoots, `dia`) {
			if err != nil {
				if !yield(nil, err) {
					return
				}
				continue
			}
			if file == nil {
				continue
			}
			cookieStore := &chrome.CookieStore{
				DefaultCookieStore: cookies.DefaultCookieStore{
					BrowserStr:           file.Browser,
					ProfileStr:           file.Profile,
					OSStr:                file.OS,
					IsDefaultProfileBool: file.IsDefaultProfile,
					FileNameStr:          file.Path,
				},
			}
			cookieStore.SetSafeStorage(`Dia`, ``, ``)
			if !yield(&cookies.CookieJar{CookieStore: cookieStore}, nil) {
				return
			}
		}
	}
}
