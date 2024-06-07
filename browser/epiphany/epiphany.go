package epiphany

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
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
		return cookies.ErrCookieSeq(errors.New(`cookie store is nil`))
	}
	if err := s.Open(); err != nil {
		return cookies.ErrCookieSeq(err)
	} else if s.Database == nil {
		return cookies.ErrCookieSeq(errors.New(`database is nil`))
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
			var expiry int64
			exp, err := row.Value(`expiry`)
			if err != nil {
				return err
			}
			switch v := exp.(type) {
			case int64:
				expiry = v
			case int32:
				expiry = int64(v)
			default:
				return fmt.Errorf("got unexpected value for Expires %v (type %[1]T)", expiry)
			}
			cookie.Expires = time.Unix(expiry, 0)

			// Secure
			sec, err := row.Value(`isSecure`)
			if err != nil {
				return err
			}
			secInt, okSec := sec.(int)
			if !okSec {
				return fmt.Errorf("got unexpected value for Secure %v (type %[1]T)", sec)
			}
			cookie.Secure = secInt > 0

			// HttpOnly
			ho, err := row.Value(`isHttpOnly`)
			if err != nil {
				return err
			}
			hoInt, okHO := ho.(int)
			if !okHO {
				return fmt.Errorf("got unexpected value for HttpOnly %v (type %[1]T)", ho)
			}
			cookie.HttpOnly = hoInt > 0
			cookie.Browser = s

			if !cookies.CookieFilterYield(&cookie, nil, yield, filters...) {
				return cookies.ErrYieldEnd
			}

			return nil
		}
	}
	seq := func(yield func(*kooky.Cookie, error) bool) {
		err := utils.VisitTableRows(s.Database, `moz_cookies`, map[string]string{}, visitor(yield))
		if !errors.Is(err, cookies.ErrYieldEnd) {
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
