package ie

import (
	"context"
	"os"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/ie"
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
	m := map[string]func(f *os.File, s *ie.CookieStore, browser string){
		`unknown`: func(f *os.File, s *ie.CookieStore, browser string) {
			t := &ie.TextCookieStore{}
			t.File = f
			t.FileNameStr = filename
			t.BrowserStr = browser
			s.CookieStore = t
		},
	}
	return ie.GetCookieStore(filename, `ie`, m, filters...)
}
