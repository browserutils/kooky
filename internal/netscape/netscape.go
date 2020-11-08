package netscape

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/zellyn/kooky"
)

const httpOnlyPrefix = `#HttpOnly_`

func (s *CookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.Open(); err != nil {
		return nil, err
	} else if s.File == nil {
		return nil, errors.New(`file is nil`)
	}

	cookies, isStrict, err := ReadCookies(s.File, filters...)
	s.IsStrictBool = isStrict

	return cookies, err
}

func ReadCookies(file io.Reader, filters ...kooky.Filter) (c []*kooky.Cookie, strict bool, e error) {
	// http://web.archive.org/web/20080520061150/wp.netscape.com/newsref/std/cookie_spec.html
	// https://github.com/Rob--W/cookie-manager/blob/83c04b74b79cb7768a33c4a93fbdfd04b90fa931/cookie-manager.js#L975
	// https://hg.python.org/cpython/file/5470dc81caf9/Lib/http/cookiejar.py#l1981

	if file == nil {
		return nil, false, errors.New(`file is nil`)
	}

	var ret []*kooky.Cookie

	var lineNr uint
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if lineNr == 0 && (line == `# HTTP Cookie File` || line == `# Netscape HTTP Cookie File`) {
			strict = true
		}
		// split line into fields
		sp := strings.Split(line, "\t")
		if len(sp) != 7 {
			continue
		}
		var exp int64
		if len(sp[4]) > 0 {
			e, err := strconv.ParseInt(sp[4], 10, 64)
			if err != nil {
				continue
			} else {
				exp = e
			}
		} else {
			// allow empty expiry field for uzbl's "session-cookies.txt" file
			strict = false
		}

		cookie := &kooky.Cookie{}
		switch sp[3] {
		case `TRUE`:
			cookie.Secure = true
		case `FALSE`:
		default:
			continue
		}

		// https://github.com/curl/curl/blob/curl-7_39_0/lib/cookie.c#L644
		// https://bugs.python.org/issue2190#msg233571
		// also in original Netscape cookies
		if strings.HasPrefix(sp[0], httpOnlyPrefix) {
			cookie.Domain = sp[0][len(httpOnlyPrefix):]
			cookie.HttpOnly = true
		} else {
			cookie.Domain = sp[0]
		}

		cookie.Path = sp[2]
		cookie.Name = sp[5]
		cookie.Value = strings.TrimSpace(sp[6])
		cookie.Expires = time.Unix(exp, 0)

		if !kooky.FilterCookie(cookie, filters...) {
			continue
		}

		ret = append(ret, cookie)
	}

	return ret, strict, nil
}
