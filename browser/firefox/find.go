package firefox

import (
	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/firefox"
	"github.com/browserutils/kooky/internal/firefox/find"
)

type firefoxFinder struct{}

var _ kooky.CookieStoreFinder = (*firefoxFinder)(nil)

func init() {
	kooky.RegisterFinder(`firefox`, &firefoxFinder{})
}

func (f *firefoxFinder) FindCookieStores() kooky.CookieStoreSeq {
	return firefox.CookieStoresForProfiles(find.FindFirefoxProfiles())
}
