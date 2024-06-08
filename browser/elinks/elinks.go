package elinks

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/iterx"
)

type elinksCookieStore struct {
	cookies.DefaultCookieStore
}

var _ cookies.CookieStore = (*elinksCookieStore)(nil)

func ReadCookies(ctx context.Context, filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	return cookies.SingleRead(cookieStore, filename, filters...).ReadAllCookies(ctx)
}

func TraverseCookies(filename string, filters ...kooky.Filter) kooky.CookieSeq {
	return cookies.SingleRead(cookieStore, filename, filters...)
}

func (s *elinksCookieStore) TraverseCookies(filters ...kooky.Filter) kooky.CookieSeq {
	if s == nil {
		return iterx.ErrCookieSeq(errors.New(`cookie store is nil`))
	}

	colCnt := 8
	parseLine := func(line string) (*kooky.Cookie, error) {
		sp := strings.Split(line, "\t")
		if l := len(sp); l != colCnt {
			return nil, fmt.Errorf(`has %d fields; expected are %d: %q`, l, colCnt, line)
		}
		exp, err := strconv.ParseInt(sp[5], 10, 64)
		if err != nil {
			return nil, fmt.Errorf(`Expires field is not an integer: %w`, err)
		}
		sec, err := strconv.Atoi(sp[6])
		if err != nil {
			return nil, fmt.Errorf(`Secure field is not an integer: %w`, err)
		}

		cookie := &kooky.Cookie{}
		cookie.Name = sp[0]
		cookie.Value = sp[1]
		cookie.Path = sp[3]
		cookie.Domain = sp[4]
		cookie.Expires = time.Unix(exp, 0)
		cookie.Secure = sec == 1
		cookie.Browser = s

		return cookie, nil
	}

	return func(yield func(*kooky.Cookie, error) bool) {
		if err := s.Open(); err != nil {
			yield(nil, err)
			return
		} else if s.File == nil {
			yield(nil, errors.New(`file is nil`))
			return
		}

		var lineNr int
		scanner := bufio.NewScanner(s.File)
		for scanner.Scan() {
			line := scanner.Text()
			lineNr++
			cookie, err := parseLine(line)
			if err != nil {
				err = fmt.Errorf(`row %d: `, lineNr)
			}
			if !iterx.CookieFilterYield(context.Background(), cookie, err, yield, filters...) {
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
	s := &elinksCookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `elinks`

	return cookies.NewCookieJar(s, filters...), nil
}
