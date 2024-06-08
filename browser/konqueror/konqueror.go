package konqueror

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

	"golang.org/x/text/encoding/charmap"
)

type konquerorCookieStore struct {
	cookies.DefaultCookieStore
}

var _ cookies.CookieStore = (*konquerorCookieStore)(nil)

func ReadCookies(ctx context.Context, filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	return cookies.SingleRead(cookieStore, filename, filters...).ReadAllCookies(ctx)
}

func TraverseCookies(filename string, filters ...kooky.Filter) kooky.CookieSeq {
	return cookies.SingleRead(cookieStore, filename, filters...)
}

func (s *konquerorCookieStore) TraverseCookies(filters ...kooky.Filter) kooky.CookieSeq {
	if s == nil {
		return iterx.ErrCookieSeq(errors.New(`cookie store is nil`))
	}

	return func(yield func(*kooky.Cookie, error) bool) {
		if err := s.Open(); err != nil {
			yield(nil, err)
		} else if s.File == nil {
			yield(nil, errors.New(`file is nil`))
		}

		latin1 := charmap.ISO8859_1.NewDecoder().Reader(s.File)

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
			// the Domain field is empty if the cookie is not for subdomains.
			// in that case the domain string does not start with '.' disallowing subdomains.
			cookie.Domain = sp[0] // fallback

			// DOMAIN
			sp = strings.SplitN(strings.TrimLeft(sp[1], ` `), `"`, 3)
			if len(sp) != 3 {
				continue
			}
			if len(sp[0]) != 0 {
				// Domain field is not quoted
				continue
			}
			if len(sp[1]) > 0 {
				// regular domain string starting with '.' allowing subdomains
				cookie.Domain = sp[1]
			}

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
	s := &konquerorCookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `konqueror`

	return cookies.NewCookieJar(s, filters...), nil
}

/*
// TODO
cookie file starts with "# KDE Cookie File v2"
-> there is probably a v1 format // TODO
*/
