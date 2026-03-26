//go:build darwin && !ios

package arc

import (
	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/chrome"
	chromefind "github.com/browserutils/kooky/internal/chrome/find"
	"github.com/browserutils/kooky/internal/cookies"
)

type arcFinder struct{}

var _ kooky.CookieStoreFinder = (*arcFinder)(nil)

func init() {
	kooky.RegisterFinder(`arc`, &arcFinder{})
}

func (f *arcFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		for file, err := range chromefind.FindCookieStoreFiles(arcChromiumRoots, `arc`) {
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
			cookieStore.SetSafeStorage(`Arc`, ``, ``)
			if !yield(&cookies.CookieJar{CookieStore: cookieStore}, nil) {
				return
			}
		}
	}
}
