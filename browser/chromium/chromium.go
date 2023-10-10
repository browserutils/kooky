package chromium

import (
	"net/http"

	"github.com/xiazemin/kooky"
	"github.com/xiazemin/kooky/internal/chrome"
	"github.com/xiazemin/kooky/internal/cookies"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s, err := cookieStore(filename, filters...)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	return s.ReadCookies(filters...)
}

// CookieJar returns an initiated http.CookieJar based on the cookies stored by
// the Chromium browser. Set cookies are memory stored and do not modify any
// browser files.
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
func CookieStore(filename string, filters ...kooky.Filter) (kooky.CookieStore, error) {
	return cookieStore(filename, filters...)
}

func cookieStore(filename string, filters ...kooky.Filter) (*cookies.CookieJar, error) {
	s := &chrome.CookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `chromium`

	return &cookies.CookieJar{CookieStore: s}, nil
}
