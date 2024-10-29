package ie

import (
	"bufio"
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/xiazemin/kooky"
	"github.com/xiazemin/kooky/internal/cookies"
	"github.com/xiazemin/kooky/internal/iterx"
	"github.com/xiazemin/kooky/internal/timex"
)

type TextCookieStore struct {
	cookies.DefaultCookieStore
}

var _ cookies.CookieStore = (*TextCookieStore)(nil)

func (s *TextCookieStore) TraverseCookies(filters ...kooky.Filter) kooky.CookieSeq {
	if s == nil {
		return iterx.ErrCookieSeq(errors.New(`cookie store is nil`))
	}
	return func(yield func(*kooky.Cookie, error) bool) {
		if err := s.Open(); err != nil {
			yield(nil, err)
			return
		} else if s.File == nil {
			yield(nil, errors.New(`file is nil`))
			return
		}
		fi, err := s.File.Stat()
		if err != nil {
			yield(nil, err)
			return
		}
		if fi.IsDir() {
			// TODO read directory content - *.txt
			yield(nil, errors.New(`file is a directory`))
			return
		}
		var (
			lineNr             int
			expLeast, crtLeast uint64
			cookie             *kooky.Cookie
			errCookie          error
		)
		scanner := bufio.NewScanner(s.File)
		for scanner.Scan() {
			lineNr = lineNr%9 + 1
			line := scanner.Text()
			if errCookie != nil && lineNr != 1 && lineNr != 9 {
				continue
			}
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
					errCookie = err
					continue
				}
				// TODO: is "Secure" encoded in flags?
				cookie.HttpOnly = flags&(1<<13) != 0
			case 5:
				var err error
				expLeast, err = strconv.ParseUint(line, 10, 32)
				if err != nil {
					errCookie = err
					continue
				}
			case 6:
				expMost, err := strconv.ParseUint(line, 10, 32)
				if err != nil {
					errCookie = err
					continue
				}
				cookie.Expires = timex.FromFILETIMESplit(expLeast, expMost)
			case 7:
				var err error
				crtLeast, err = strconv.ParseUint(line, 10, 32)
				if err != nil {
					errCookie = err
					continue
				}
			case 8:
				crtMost, err := strconv.ParseUint(line, 10, 32)
				if err != nil {
					errCookie = err
					continue
				}
				cookie.Creation = timex.FromFILETIMESplit(crtLeast, crtMost)
			case 9:
				// Secure (?)
				if line != `*` {
					errCookie = errors.New(`cookie record delimiter not "*"`)
				}
				if errCookie == nil {
					cookie.Browser = s
				} else {
					cookie = nil
				}
				if !iterx.CookieFilterYield(context.Background(), cookie, errCookie, yield, filters...) {
					return
				}
				errCookie = nil
			}
		}
	}
}
