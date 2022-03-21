package opera

import (
	"errors"
	"path/filepath"

	"github.com/zellyn/kooky"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s := &operaCookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `opera`

	defer s.Close()

	return s.ReadCookies(filters...)
}

func (s *operaCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.Open(); err != nil {
		return nil, err
	}

	switch filepath.Base(s.FileNameStr) {
	case `cookies4.dat`:
		if s.File == nil {
			return nil, errors.New(`file is nil`)
		}
		return s.readPrestoCookies(filters...)
	case `Cookies`:
		fallthrough
	default:
		if s.Database == nil {
			return nil, errors.New(`database is nil`)
		}
		// Chrome sqlite format
		return s.readBlinkCookies(filters...)
	}
}

// "cookies4.dat" format
func (s *operaCookieStore) readPrestoCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	// https://web.archive.org/web/20100303220606/www.opera.com/docs/fileformats#cookies
	// https://stackoverflow.com/a/12223897
	// https://www.codeproject.com/Articles/330142/Cookie-Quest-A-Quest-to-Read-Cookies-from-Four-Pop#Opera4
	// http://users.westelcom.com/jsegur/O4FE.HTM#TS1
	//
	// TODO: Presto cookiestore filenames: "cookies4.dat", "cookies4.new", "cookies4.old", "cookies.dat", `C:\klient\dcookie.txt`

	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}

	return nil, errors.New(`not implemented`)
}

// https://gist.github.com/pich4ya/5918c629b3bf3c42e696f07db354d80b
// 'Login Data' sqlite file
// SELECT origin_url, username_value, password_value FROM logins

func (s *operaCookieStore) readBlinkCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}

	cookies, err := s.CookieStore.ReadCookies(filters...)

	return cookies, err
}
