package w3m

import (
	"bufio"
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/iterx"
)

type w3mCookieStore struct {
	cookies.DefaultCookieStore
}

var _ cookies.CookieStore = (*w3mCookieStore)(nil)

func ReadCookies(ctx context.Context, filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	return cookies.SingleRead(cookieStore, filename, filters...).ReadAllCookies(ctx)
}

func TraverseCookies(filename string, filters ...kooky.Filter) kooky.CookieSeq {
	return cookies.SingleRead(cookieStore, filename, filters...)
}

func (s *w3mCookieStore) TraverseCookies(filters ...kooky.Filter) kooky.CookieSeq {
	// cookie.c: void save_cookies(void){}
	// https://github.com/tats/w3m/blob/169789b1480710712d587d5859fab9d93eb952a2/cookie.c#L429

	if s == nil {
		return iterx.ErrCookieSeq(errors.New(`cookie store is nil`))
	}
	if err := s.Open(); err != nil {
		return iterx.ErrCookieSeq(err)
	} else if s.File == nil {
		return iterx.ErrCookieSeq(errors.New(`file is nil`))
	}

	return func(yield func(*kooky.Cookie, error) bool) {
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
			cookie.Browser = s

			if !iterx.CookieFilterYield(context.Background(), cookie, nil, yield, filters...) {
				return
			}
		}
	}
}

// CookieStore has to be closed with CookieStore.Close() after use.
func CookieStore(filename string, filters ...kooky.Filter) (kooky.CookieStore, error) {
	return cookieStore(filename, filters...)
}

func cookieStore(filename string, filters ...kooky.Filter) (*cookies.CookieJar, error) {
	s := &w3mCookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `w3m`

	return cookies.NewCookieJar(s, filters...), nil
}

// TODO:
// change behaviour based on version (8th field) - current implementation is for version 0
// implement old v0 format recognition and use netscape parser
