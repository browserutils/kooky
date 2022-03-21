package chrome

import (
	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/chrome"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s := &chrome.CookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `chrome`

	defer s.Close()

	return s.ReadCookies(filters...)
}
