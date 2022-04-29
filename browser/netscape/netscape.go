package netscape

import (
	"net/http"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/cookies"
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

// CookieJar returns an initiated http.CookieJar based on the cookies stored by
// the Netscape browser. Set cookies are memory stored and do not modify any
// browser files.
//
func CookieJar(filename string, filters ...kooky.Filter) (http.CookieJar, error) {
	j, err := cookieStore(filename, filters...)
	if err != nil {
		return nil, err
	}
	defer j.Close()
	if err := j.InitJar(); err != nil {
		return nil, err
	}
	return j, nil
}

// CookieStore has to be closed with CookieStore.Close() after use.
//
func CookieStore(filename string, filters ...kooky.Filter) (kooky.CookieStore, error) {
	return cookieStore(filename, filters...)
}

func cookieStore(filename string, filters ...kooky.Filter) (*cookies.CookieJar, error) {
	s := &netscape.CookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `netscape`

	return &cookies.CookieJar{CookieStore: s}, nil
}
