package netscape

import (
	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/firefox/find"
	"github.com/browserutils/kooky/internal/netscape"
)

type netscapeFinder struct{}

var _ kooky.CookieStoreFinder = (*netscapeFinder)(nil)

func init() {
	kooky.RegisterFinder(`netscape`, &netscapeFinder{})
}

func (f *netscapeFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		for file, err := range find.FindCookieStoreFiles(netscapeRoots, `netscape`, `cookies.txt`) {
			if err != nil {
				if !yield(nil, err) {
					return
				}
				continue
			}
			st := &cookies.CookieJar{
				CookieStore: &netscape.CookieStore{
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
