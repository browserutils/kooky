//go:build js

package website

import (
	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
)

type websiteFinder struct{}

var _ kooky.CookieStoreFinder = (*websiteFinder)(nil)

func init() {
	kooky.RegisterFinder(browserName, &websiteFinder{})
}

func (f *websiteFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		s := newStore()
		st := &cookies.CookieJar{CookieStore: s}
		if !yield(st, nil) {
			return
		}
	}
}
