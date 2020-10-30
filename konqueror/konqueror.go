package konqueror

import (
	"bufio"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/zellyn/kooky"

	"golang.org/x/text/encoding/charmap"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s := &konquerorCookieStore{filename: filename}
	defer s.Close()

	return s.ReadCookies(filters...)
}

func (s *konquerorCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.open(); err != nil {
		return nil, err
	} else if s.file == nil {
		return nil, errors.New(`file is nil`)
	}

	var ret []*kooky.Cookie

	latin1 := charmap.ISO8859_1.NewDecoder().Reader(s.file)

	scanner := bufio.NewScanner(latin1)
	for scanner.Scan() {
		line := scanner.Text()
		// lines should be at least 97 characters long
		// skip comments and domain section headers
		if len(line) == 0 || line[0] == '#' || line[0] == '[' {
			continue
		}

		// Host Domain Path Expires Prot Name Sec Value

		// HOST
		sp := strings.SplitN(line, ` `, 2)
		if len(sp) != 2 {
			continue
		}
		cookie := &kooky.Cookie{}
		// Domain field is sometimes empty
		// Host and Domain fields seem to be the same otherwise // TODO
		cookie.Domain = sp[0]

		// DOMAIN
		sp = strings.SplitN(strings.TrimLeft(sp[1], ` `), `"`, 3)
		if len(sp) != 3 {
			continue
		}
		if len(sp[0]) != 0 {
			// Domain field is not quoted
			continue
		}
		// ignore Domain field (it can be empty)

		// PATH
		sp = strings.SplitN(strings.TrimLeft(sp[2], ` `), `"`, 3)
		if len(sp) != 3 || len(sp[0]) != 0 {
			// Path field is not quoted (if sp[0] empty)
			continue
		}
		cookie.Path = sp[1]

		// EXPIRES
		sp = strings.SplitN(strings.TrimLeft(sp[2], ` `), ` `, 2)
		if len(sp) != 2 {
			continue
		}
		exp, err := strconv.ParseInt(sp[0], 10, 64)
		if err != nil {
			continue
		}

		// PROT
		// skip
		sp = strings.SplitN(strings.TrimLeft(sp[1], ` `), ` `, 2)
		if len(sp) != 2 {
			continue
		}
		if _, err := strconv.Atoi(sp[0]); err != nil {
			continue
		}

		// NAME
		sp = strings.SplitN(strings.TrimLeft(sp[1], ` `), ` `, 2)
		if len(sp) != 2 {
			continue
		}
		cookie.Name = sp[0]

		// SEC
		sp = strings.SplitN(strings.TrimLeft(sp[1], ` `), ` `, 2)
		if len(sp) != 2 {
			continue
		}
		sec, err := strconv.Atoi(sp[0])
		if err != nil {
			continue
		}

		// VALUE
		cookie.Value = strings.Trim(sp[1], ` `)

		cookie.Expires = time.Unix(exp, 0)

		const (
			secure          int = 1
			httpOnly        int = 2
			hasExplicitPath int = 4
			emptyName       int = 8
		)
		cookie.Secure = sec&secure != 0
		cookie.HttpOnly = sec&httpOnly != 0

		if !kooky.FilterCookie(cookie, filters...) {
			continue
		}

		ret = append(ret, cookie)
	}
	return ret, nil
}
