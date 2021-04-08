package netscape

import (
	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/netscape"
)

// This ReadCookies() function returns an additional boolean "strict" telling
// if the file adheres to the netscape cookies.txt format
func ReadCookies(filename string, filters ...kooky.Filter) (c []*kooky.Cookie, strict bool, e error) {
	s := &netscape.CookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `netscape`

	defer s.Close()

	cookies, err := s.ReadCookies(filters...)

	return cookies, s.IsStrict(), err
}
