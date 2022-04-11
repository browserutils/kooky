package w3m

import (
	"bufio"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/cookies"
)

type w3mCookieStore struct {
	cookies.DefaultCookieStore
}

var _ cookies.CookieStore = (*w3mCookieStore)(nil)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s, err := cookieStore(filename, filters...)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	return s.ReadCookies(filters...)
}

func (s *w3mCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	// cookie.c: void save_cookies(void){}
	// https://github.com/tats/w3m/blob/169789b1480710712d587d5859fab9d93eb952a2/cookie.c#L429

	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.Open(); err != nil {
		return nil, err
	} else if s.File == nil {
		return nil, errors.New(`file is nil`)
	}

	var ret []*kooky.Cookie

	scanner := bufio.NewScanner(s.File)
	for scanner.Scan() {
		// split line into fields
		sp := strings.Split(scanner.Text(), "\t")
		if len(sp) != 11 {
			continue
		}
		exp, err := strconv.ParseInt(sp[3], 10, 64)
		if err != nil {
			continue
		}
		bitFlag, err := strconv.Atoi(sp[6])
		if err != nil {
			continue
		}

		// #defined in "fm.h"
		const (
			// cooUse      int =  1 // COO_USE
			cooSecure int = 2 // COO_SECURE
			// cooDomain   int =  4 // COO_DOMAIN
			// cooPath     int =  8 // COO_PATH
			// cooDiscard  int = 16 // COO_DISCARD
			// cooOverride int = 32 // COO_OVERRIDE - user override of security checks
		)

		cookie := &kooky.Cookie{}
		cookie.Name = sp[1]
		cookie.Value = sp[2]
		cookie.Path = sp[5]
		cookie.Domain = sp[4]
		cookie.Expires = time.Unix(exp, 0)
		cookie.Secure = bitFlag&cooSecure != 0
		// sp[6] // state management specification version
		// sp[7] // port list

		if !kooky.FilterCookie(cookie, filters...) {
			continue
		}

		ret = append(ret, cookie)
	}
	return ret, nil
}

// CookieJar returns an initiated http.CookieJar based on the cookies stored by
// the w3m browser. Set cookies are memory stored and do not modify any
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
	s := &w3mCookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `w3m`

	return &cookies.CookieJar{CookieStore: s}, nil
}

// TODO:
// change behaviour based on version (8th field) - current implementation is for version 0
// implement old v0 format recognition and use netscape parser
