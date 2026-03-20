//go:build js

package website

import (
	"context"
	"errors"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/iterx"
	"github.com/browserutils/kooky/internal/website"
)

const browserName = `website`

type websiteCookieStore struct {
	cookies.DefaultCookieStore
}

var _ cookies.CookieStore = (*websiteCookieStore)(nil)

func newStore() *websiteCookieStore {
	s := &websiteCookieStore{}
	s.BrowserStr = browserName
	s.IsDefaultProfileBool = true
	return s
}

func ReadCookies(ctx context.Context, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	return cookies.ReadCookiesClose(newStore(), filters...).ReadAllCookies(ctx)
}

func TraverseCookies(filters ...kooky.Filter) kooky.CookieSeq {
	return cookies.ReadCookiesClose(newStore(), filters...)
}

// CookieStore has to be closed with CookieStore.Close() after use.
func CookieStore(filters ...kooky.Filter) (kooky.CookieStore, error) {
	return cookieStoreFunc(filters...)
}

func cookieStoreFunc(filters ...kooky.Filter) (*cookies.CookieJar, error) {
	return cookies.NewCookieJar(newStore(), filters...), nil
}

func (s *websiteCookieStore) TraverseCookies(filters ...kooky.Filter) kooky.CookieSeq {
	if s == nil {
		return iterx.ErrCookieSeq(errors.New(`cookie store is nil`))
	}
	return func(yield func(*kooky.Cookie, error) bool) {
		for c, err := range website.TraverseCookies(s, nil, filters...) {
			if err != nil {
				if !yield(nil, err) {
					return
				}
				continue
			}
			if !yield(c.Cookie, nil) {
				return
			}
		}
	}
}
