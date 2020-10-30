package opera

import (
	"errors"
	"path/filepath"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/chrome"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s := &operaCookieStore{
		filename: filename,
		browser:  `opera`,
	}
	defer s.Close()

	return s.ReadCookies(filters...)
}

func (s *operaCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.open(); err != nil {
		return nil, err
	}

	switch filepath.Base(s.filename) {
	case `cookies4.dat`:
		if s.file == nil {
			return nil, errors.New(`file is nil`)
		}
		return s.readPrestoCookies(filters...)
	case `Cookies`:
		fallthrough
	default:
		if s.database == nil {
			return nil, errors.New(`database is nil`)
		}
		// Chrome sqlite format
		/*
			// TODO decryption fails (linux)
			cookies, err := ReadChromeCookies(filename, ``, ``, time.Time{})
			if err != nil {
				return nil, err
			}
			return FilterCookies(cookies, filters...), nil
		*/
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

	return nil, errors.New(`not implemented`)
}

// https://gist.github.com/pich4ya/5918c629b3bf3c42e696f07db354d80b
// 'Login Data' sqlite file
// SELECT origin_url, username_value, password_value FROM logins

func (s *operaCookieStore) readBlinkCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	cookieStore := &chrome.CookieStore{
		Filename:         s.filename,
		Database:         s.database,
		KeyringPassword:  s.keyringPassword,
		Password:         s.password,
		OS:               s.os,
		Browser:          s.browser,
		Profile:          s.profile,
		IsDefaultProfile: s.isDefaultProfile,
		DecryptionMethod: s.decryptionMethod,
	}

	cookies, err := cookieStore.ReadCookies(filters...)

	s.filename = cookieStore.Filename
	s.database = cookieStore.Database
	s.keyringPassword = cookieStore.KeyringPassword
	s.password = cookieStore.Password
	s.os = cookieStore.OS
	s.browser = cookieStore.Browser
	s.profile = cookieStore.Profile
	s.isDefaultProfile = cookieStore.IsDefaultProfile
	s.decryptionMethod = cookieStore.DecryptionMethod

	return cookies, err
}
