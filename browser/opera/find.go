package opera

import (
	"path/filepath"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/chrome"
	chromefind "github.com/browserutils/kooky/internal/chrome/find"
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

		for file, err := range chromefind.FindCookieStoreFiles(operaBlinkRoots, `opera`) {
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
			cookieStore.SetSafeStorage(operaSafeStorageAccount, ``, ``)
			st := &cookies.CookieJar{
				CookieStore: &operaCookieStore{
					CookieStore: cookieStore,
				},
			}
			if !yield(st, nil) {
				return
			}
		}
	}
}
