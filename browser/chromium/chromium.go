package chromium

import (
	"context"

	"github.com/xiazemin/kooky"
	"github.com/xiazemin/kooky/internal/chrome"
	"github.com/xiazemin/kooky/internal/cookies"
)

func ReadCookies(ctx context.Context, filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	return cookies.SingleRead(cookieStore, filename, filters...).ReadAllCookies(ctx)
}

func TraverseCookies(filename string, filters ...kooky.Filter) kooky.CookieSeq {
	return cookies.SingleRead(cookieStore, filename, filters...)
}

// CookieStore has to be closed with CookieStore.Close() after use.
func CookieStore(filename string, filters ...kooky.Filter) (kooky.CookieStore, error) {
	return cookieStore(filename, filters...)
}

func cookieStore(filename string, filters ...kooky.Filter) (*cookies.CookieJar, error) {
	s := &chrome.CookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `chromium`

	return cookies.NewCookieJar(s, filters...), nil
}
