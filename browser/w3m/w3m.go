package w3m

import (
	"bufio"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal"
)

type w3mCookieStore struct {
	internal.DefaultCookieStore
}

var _ kooky.CookieStore = (*w3mCookieStore)(nil)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s := &w3mCookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `w3m`

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

// TODO:
// change behaviour based on version (8th field) - current implementation is for version 0
// implement old v0 format recognition and use netscape parser
