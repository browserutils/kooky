package netscape

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/iterx"
)

const httpOnlyPrefix = `#HttpOnly_`

func (s *CookieStore) TraverseCookies(filters ...kooky.Filter) kooky.CookieSeq {
	if s == nil {
		return iterx.ErrCookieSeq(errors.New(`cookie store is nil`))
	}
	if err := s.Open(); err != nil {
		return iterx.ErrCookieSeq(err)
	}
	if s.File == nil {
		return iterx.ErrCookieSeq(errors.New(`file is nil`))
	}

	seq, str := TraverseCookies(s.File, s, filters...)
	s.isStrict = str

	return seq
}

func TraverseCookies(file io.Reader, bi kooky.BrowserInfo, filters ...kooky.Filter) (_ kooky.CookieSeq, isStrict func() bool) {
	// http://web.archive.org/web/20080520061150/wp.netscape.com/newsref/std/cookie_spec.html
	// https://github.com/Rob--W/cookie-manager/blob/83c04b74b79cb7768a33c4a93fbdfd04b90fa931/cookie-manager.js#L975
	// https://hg.python.org/cpython/file/5470dc81caf9/Lib/http/cookiejar.py#l1981

	parseLine := func(line string, lineNr int, strPtr *bool, yield func(*kooky.Cookie, error) bool) bool {
		// split line into fields
		sp := strings.Split(line, "\t")
		colCnt := 7
		if l := len(sp); l != colCnt {
			if len(line) == 0 || strings.HasPrefix(line, `#`) {
				// comment
				return true // continue
			}
			return yield(nil, fmt.Errorf(`row %d: has %d fields; expected are %d: %q`, lineNr+1, l, colCnt, line))
		}
		var exp int64
		if len(sp[4]) > 0 {
			e, err := strconv.ParseInt(sp[4], 10, 64)
			if err != nil {
				return yield(nil, fmt.Errorf(`row %d: Expires field is not an integer: %w`, lineNr+1, err))
			} else {
				exp = e
			}
		} else {
			// allow empty expiry field for uzbl's "session-cookies.txt" file
			if strPtr != nil && *strPtr {
				*strPtr = false
				if !yield(nil, ErrNotStrict) {
					return false
				}
			}
		}

		cookie := &kooky.Cookie{}
		switch sp[3] {
		case `TRUE`:
			cookie.Secure = true
		case `FALSE`:
		default:
			return yield(nil, fmt.Errorf(`row %d: Secure field is not a bool`, lineNr+1))
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
		cookie.Browser = bi

		return iterx.CookieFilterYield(context.Background(), cookie, nil, yield, filters...)
	}

	var strict bool
	done := make(chan struct{}, 1)
	isStrict = func() bool {
		<-done
		return strict
	}

	seq := func(yield func(*kooky.Cookie, error) bool) {
		defer func() { done <- struct{}{} }()

		if file == nil {
			yield(nil, errors.New(`file is nil`))
			return
		}

		var lineNr int
		scanner := bufio.NewScanner(file)
		if !scanner.Scan() {
			yield(nil, errors.New(`file has no lines`))
			return
		}
		line := scanner.Text()
		if line == `# HTTP Cookie File` || line == `# Netscape HTTP Cookie File` {
			strict = true
		} else {
			if !yield(nil, ErrNotStrict) {
				return
			}
		}
		lineNr++
		if !parseLine(line, lineNr, &strict, yield) {
			return
		}
		for scanner.Scan() {
			line := scanner.Text()
			lineNr++
			if !parseLine(line, lineNr, &strict, yield) {
				return
			}
		}
	}

	return seq, isStrict
}

var ErrNotStrict = errors.New(`netscape cookie file: file format not strictly followed`)
