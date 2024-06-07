package netscape

import (
	"github.com/browserutils/kooky/internal/cookies"
)

type CookieStore struct {
	cookies.DefaultCookieStore
	isStrict func() bool
}

// strict netscape cookies.txt format
func (s *CookieStore) IsStrict() bool {
	if s == nil {
		return false
	}
	if s.isStrict == nil {
		_, s.isStrict = TraverseCookies(s.File, s)
	}
	return s.isStrict()
}

var _ cookies.CookieStore = (*CookieStore)(nil)
