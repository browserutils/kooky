// Edge Browser shared text cookies with IE up to Fall Creators Update 1709 of Windows 10.
// After that cookies were store in the ESE database WebCacheV01.dat up to Edge v44.
// Currently cookies are stored in the same way as the Chrome browser stores them.

package edge

import (
	"context"
	"os"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/chrome"
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
		`sqlite`: func(f *os.File, s *ie.CookieStore, browser string) {
			f.Close()
			c := &chrome.CookieStore{}
			c.FileNameStr = filename
			c.BrowserStr = `edge`
			s.CookieStore = c
		},
	}
	return ie.GetCookieStore(filename, `edge`, m, filters...)
}
