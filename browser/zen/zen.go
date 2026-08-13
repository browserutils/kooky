// Zen (https://zen-browser.app) is a Firefox fork that keeps its own
// profile tree, separate from Firefox's, but reuses Firefox's cookies.sqlite
// and sessionstore formats verbatim.
package zen

import (
	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/firefox"
	"github.com/browserutils/kooky/internal/firefox/find"
)

type zenFinder struct{}

var _ kooky.CookieStoreFinder = (*zenFinder)(nil)

func init() {
	kooky.RegisterFinder(`zen`, &zenFinder{})
}

func (f *zenFinder) FindCookieStores() kooky.CookieStoreSeq {
	return firefox.CookieStoresForProfiles(find.FindZenProfiles())
}
