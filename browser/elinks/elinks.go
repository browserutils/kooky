package elinks

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

type elinksCookieStore struct {
	cookies.DefaultCookieStore
}

var _ cookies.CookieStore = (*elinksCookieStore)(nil)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s, err := cookieStore(filename, filters...)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	return s.ReadCookies(filters...)
}

func (s *elinksCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
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
		line := scanner.Text()
		sp := strings.Split(line, "\t")
		if len(sp) != 8 {
			continue
		}
		exp, err := strconv.ParseInt(sp[5], 10, 64)
		if err != nil {
			continue
		}
		sec, err := strconv.Atoi(sp[6])
		if err != nil {
			continue
		}

		cookie := &kooky.Cookie{}

		cookie.Name = sp[0]
		cookie.Value = sp[1]
		cookie.Path = sp[3]
		cookie.Domain = sp[4]
		cookie.Expires = time.Unix(exp, 0)
		cookie.Secure = sec == 1

		if !kooky.FilterCookie(cookie, filters...) {
			continue
		}

		ret = append(ret, cookie)
	}
	return ret, nil
}

// CookieJar returns an initiated http.CookieJar based on the cookies stored by
// the ELinks browser. Set cookies are memory stored and do not modify any
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
	s := &elinksCookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `elinks`

	return &cookies.CookieJar{CookieStore: s}, nil
}
