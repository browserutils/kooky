package dillo

import (
	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/netscape"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s := &netscape.CookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `dillo`

	defer s.Close()

	return s.ReadCookies(filters...)
}
