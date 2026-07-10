package vivaldi

import (
	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/chrome"
	chromefind "github.com/browserutils/kooky/internal/chrome/find"
	"github.com/browserutils/kooky/internal/cookies"
)

type vivaldiFinder struct{}

var _ kooky.CookieStoreFinder = (*vivaldiFinder)(nil)

func init() {
	kooky.RegisterFinder(`vivaldi`, &vivaldiFinder{})
}

func (f *vivaldiFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		for file, err := range chromefind.FindCookieStoreFiles(vivaldiRoots, `vivaldi`) {
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
			// macOS derives "Vivaldi Safe Storage" from the browser name; Linux
			// (KWallet/portal) needs the xdg-desktop-portal application ID.
			cookieStore.SetPortalAppID(`com.vivaldi.Vivaldi`)
			if !yield(&cookies.CookieJar{CookieStore: cookieStore}, nil) {
				return
			}
		}
	}
}
