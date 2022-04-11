// Edge Browser shared text cookies with IE up to Fall Creators Update 1709 of Windows 10.
// After that cookies were store in the ESE database WebCacheV01.dat up to Edge v44.
// Currently cookies are stored in the same way as the Chrome browser stores them.

package edge

import (
	"net/http"
	"os"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/chrome"
	"github.com/zellyn/kooky/internal/cookies"
	"github.com/zellyn/kooky/internal/ie"
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
// the Edge browser. Set cookies are memory stored and do not modify any
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
