package opera

import (
	"errors"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/chrome"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/iterx"
	"github.com/browserutils/kooky/internal/utils"
)

func TraverseCookies(filename string, filters ...kooky.Filter) kooky.CookieSeq {
	return cookies.SingleRead(cookieStore, filename, filters...)
}

func (s *operaCookieStore) TraverseCookies(filters ...kooky.Filter) kooky.CookieSeq {
	if s == nil {
		return iterx.ErrCookieSeq(errors.New(`cookie store is nil`))
	}
	return s.CookieStore.TraverseCookies(filters...)
}

// CookieStore has to be closed with CookieStore.Close() after use.
func CookieStore(filename string, filters ...kooky.Filter) (kooky.CookieStore, error) {
	return cookieStore(filename, filters...)
}

func cookieStore(filename string, filters ...kooky.Filter) (*cookies.CookieJar, error) {
	s := &operaCookieStore{}

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

	return cookies.NewCookieJar(s, filters...), nil
}
