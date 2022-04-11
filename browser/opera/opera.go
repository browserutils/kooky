package opera

import (
	"errors"
	"net/http"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/chrome"
	"github.com/zellyn/kooky/internal/cookies"
	"github.com/zellyn/kooky/internal/utils"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s, err := cookieStore(filename, filters...)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	return s.ReadCookies(filters...)
}

func (s *operaCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	return s.CookieStore.ReadCookies(filters...)
}

// CookieJar returns an initiated http.CookieJar based on the cookies stored by
// the Opera browser. Set cookies are memory stored and do not modify any
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
	var s operaCookieStore

	f, typ, err := utils.DetectFileType(filename)
	if err != nil {
		return nil, err
	}
	switch typ {
	case `opera_cookies4_1.0`: // TODO `presto`
		p := &operaPrestoCookieStore{}
		p.File = f
		p.FileNameStr = filename
		p.BrowserStr = `opera`
		s.CookieStore = p
	case `sqlite`:
		// based on Chrome browser
		// TODO: implement in internal/chrome
		//
		// https://gist.github.com/pich4ya/5918c629b3bf3c42e696f07db354d80b
		// 'Login Data' sqlite file
		// SELECT origin_url, username_value, password_value FROM logins
		// func (s *operaCookieStore) readBlinkCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
		f.Close()
		c := &chrome.CookieStore{}
		c.FileNameStr = filename
		c.BrowserStr = `opera`
		s.CookieStore = c
	default:
		f.Close()
		return nil, errors.New(`unknown file type`)
	}

	return &cookies.CookieJar{CookieStore: &s}, nil
}
