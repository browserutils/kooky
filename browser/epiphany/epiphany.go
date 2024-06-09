package epiphany

import (
	"context"
	"errors"
	"time"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/iterx"
	"github.com/browserutils/kooky/internal/utils"
)

func ReadCookies(ctx context.Context, filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	return cookies.SingleRead(cookieStore, filename, filters...).ReadAllCookies(ctx)
}

func TraverseCookies(filename string, filters ...kooky.Filter) kooky.CookieSeq {
	return cookies.SingleRead(cookieStore, filename, filters...)
}

func (s *epiphanyCookieStore) TraverseCookies(filters ...kooky.Filter) kooky.CookieSeq {
	if s == nil {
		return iterx.ErrCookieSeq(errors.New(`cookie store is nil`))
	}
	if err := s.Open(); err != nil {
		return iterx.ErrCookieSeq(err)
	} else if s.Database == nil {
		return iterx.ErrCookieSeq(errors.New(`database is nil`))
	}

	// Epiphany originally used a Mozilla Gecko backend but later switched to WebKit.
	// For possible deviations from the firefox database layout
	// it might be better not to depend on the firefox implementation.

	visitor := func(yield func(*kooky.Cookie, error) bool) func(rowId *int64, row utils.TableRow) error {
		return func(rowId *int64, row utils.TableRow) error {
			cookie := kooky.Cookie{}
			var err error

			// Name
			cookie.Name, err = row.String(`name`)
			if err != nil {
				return err
			}

			// Value
			cookie.Value, err = row.String(`value`)
			if err != nil {
				return err
			}

			// Host
			cookie.Domain, err = row.String(`host`)
			if err != nil {
				return err
			}

			// Path
			cookie.Path, err = row.String(`path`)
			if err != nil {
				return err
			}

			// Expires
			expiry, err := utils.ValueOrFallback[int64](row, `expiry`, 0, true)
			if err != nil {
				return err
			}
			cookie.Expires = time.Unix(expiry, 0)

			// Secure
			cookie.Secure, err = row.Bool(`isSecure`)
			if err != nil {
				return err
			}

			// HttpOnly
			cookie.HttpOnly, err = row.Bool(`isHttpOnly`)
			if err != nil {
				return err
			}
			cookie.Browser = s

			if !iterx.CookieFilterYield(context.Background(), &cookie, nil, yield, filters...) {
				return iterx.ErrYieldEnd
			}

			return nil
		}
	}
	seq := func(yield func(*kooky.Cookie, error) bool) {
		err := utils.VisitTableRows(s.Database, `moz_cookies`, map[string]string{}, visitor(yield))
		if !errors.Is(err, iterx.ErrYieldEnd) {
			yield(nil, err)
		}
	}

	return seq
}

// CookieStore has to be closed with CookieStore.Close() after use.
func CookieStore(filename string, filters ...kooky.Filter) (kooky.CookieStore, error) {
	return cookieStore(filename, filters...)
}

func cookieStore(filename string, filters ...kooky.Filter) (*cookies.CookieJar, error) {
	s := &epiphanyCookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `epiphany`

	return cookies.NewCookieJar(s, filters...), nil
}
