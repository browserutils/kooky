package ie

import (
	"net/http"
	"os"

	"github.com/zellyn/kooky"
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
// the Internet Explorer browser. Set cookies are memory stored and do not modify any
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
