package netscape

import (
	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal"
)

type CookieStore struct {
	internal.DefaultCookieStore
	IsStrictBool bool
}

// strict netscape cookies.txt format
func (s *CookieStore) IsStrict() bool {
	return s != nil && s.IsStrictBool
}

var _ kooky.CookieStore = (*CookieStore)(nil)
