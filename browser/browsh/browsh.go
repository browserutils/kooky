// Browsh Browser
package browsh

import (
	"context"

	"github.com/xiazemin/kooky"
	"github.com/xiazemin/kooky/internal/cookies"
	"github.com/xiazemin/kooky/internal/firefox"
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
	s := &firefox.CookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `browsh`

	return cookies.NewCookieJar(s, filters...), nil
}
