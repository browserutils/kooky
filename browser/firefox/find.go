package firefox

import (
	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/firefox"
	"github.com/browserutils/kooky/internal/firefox/find"
)

type firefoxFinder struct{}

var _ kooky.CookieStoreFinder = (*firefoxFinder)(nil)

func init() {
	kooky.RegisterFinder(`firefox`, &firefoxFinder{})
}

func (f *firefoxFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		for file, err := range find.FindFirefoxCookieStoreFiles() {
			if err != nil {
				if !yield(nil, err) {
					return
				}
			}
			st := &cookies.CookieJar{
				CookieStore: &firefox.CookieStore{
					DefaultCookieStore: cookies.DefaultCookieStore{
						BrowserStr:           file.Browser,
						ProfileStr:           file.Profile,
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
