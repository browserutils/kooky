package elinks

import (
	"bufio"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/zellyn/kooky"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s := &elinksCookieStore{filename: filename}
	defer s.Close()

	return s.ReadCookies(filters...)
}

func (s *elinksCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.open(); err != nil {
		return nil, err
	} else if s.file == nil {
		return nil, errors.New(`file is nil`)
	}

	var ret []*kooky.Cookie

	scanner := bufio.NewScanner(s.file)
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

		// src/cookies/cookies.c
		// enum { NAME = 0, VALUE, SERVER, PATH, DOMAIN, EXPIRES, SECURE, MEMBERS };
		cookie.Name = sp[0]
		cookie.Value = sp[1]
		cookie.Path = sp[3]
		cookie.Domain = sp[4]
		cookie.Expires = time.Unix(exp, 0)
		cookie.Secure = sec > 1

		if !kooky.FilterCookie(cookie, filters...) {
			continue
		}

		ret = append(ret, cookie)
	}
	return ret, nil
}
