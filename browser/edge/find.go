package edge

import (
	"runtime"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/chrome"
	chromefind "github.com/browserutils/kooky/internal/chrome/find"
	"github.com/browserutils/kooky/internal/cookies"
)

type edgeFinder struct{}

var _ kooky.CookieStoreFinder = (*edgeFinder)(nil)

func init() {
	kooky.RegisterFinder(`edge`, &edgeFinder{})
}

func (f *edgeFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		for file, err := range chromefind.FindCookieStoreFiles(edgeChromiumRoots, `edge`) {
			if err != nil {
				if !yield(nil, err) {
					return
				}
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
			cookieStore.SetSafeStorage(`Microsoft Edge`, ``)
			if !yield(&cookies.CookieJar{CookieStore: cookieStore}, nil) {
				return
			}
		}

		if runtime.GOOS != `windows` || edgeOldCookieStores == nil {
			return
		}
		for oldCookieStore, err := range edgeOldCookieStores { // ESE, text cookies
			if err != nil {
				if !yield(nil, err) {
					return
				}
				continue
			}
			if !yield(oldCookieStore, nil) {
				return
			}
		}
	}
}

var edgeOldCookieStores kooky.CookieStoreSeq
