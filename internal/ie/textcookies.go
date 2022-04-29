package ie

import (
	"bufio"
	"errors"
	"strconv"
	"strings"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/cookies"
	"github.com/zellyn/kooky/internal/timex"
)

type TextCookieStore struct {
	cookies.DefaultCookieStore
}

var _ cookies.CookieStore = (*TextCookieStore)(nil)

func (s *TextCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.Open(); err != nil {
		return nil, err
	} else if s.File == nil {
		return nil, errors.New(`file is nil`)
	}
	fi, err := s.File.Stat()
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		// TODO read directory content - *.txt
	}

	var (
		lineNr             int
		expLeast, crtLeast uint64
		cookie             *kooky.Cookie
		cookies            []*kooky.Cookie
	)
	scanner := bufio.NewScanner(s.File)
	for scanner.Scan() {
		lineNr = lineNr%9 + 1
		line := scanner.Text()
		switch lineNr {
		case 1:
			cookie = &kooky.Cookie{}
			cookie.Name = line
		case 2:
			cookie.Value = line
		case 3:
			sp := strings.SplitN(line, `/`, 2)
			cookie.Domain = sp[0]
			if len(sp) == 2 {
				cookie.Path = `/` + sp[1]
			} else {
				cookie.Path = `/`
			}
		case 4:
			flags, err := strconv.ParseUint(line, 10, 32)
			if err != nil {
				return nil, err
			}
			// TODO: is "Secure" encoded in flags?
			cookie.HttpOnly = flags&(1<<13) != 0
		case 5:
			var err error
			expLeast, err = strconv.ParseUint(line, 10, 32)
			if err != nil {
				return nil, err
			}
		case 6:
			expMost, err := strconv.ParseUint(line, 10, 32)
			if err != nil {
				return nil, err
			}
			cookie.Expires = timex.FromFILETIMESplit(expLeast, expMost)
		case 7:
			var err error
			crtLeast, err = strconv.ParseUint(line, 10, 32)
			if err != nil {
				return nil, err
			}
		case 8:
			crtMost, err := strconv.ParseUint(line, 10, 32)
			if err != nil {
				return nil, err
			}
			cookie.Creation = timex.FromFILETIMESplit(crtLeast, crtMost)
		case 9:
			// Secure (?)
			if line != `*` {
				return nil, errors.New(`cookie record delimiter not "*"`)
			}
			if kooky.FilterCookie(cookie, filters...) {
				cookies = append(cookies, cookie)
			}
		}
	}

	return cookies, nil
}
