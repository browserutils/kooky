package epiphany

import (
	"errors"
	"net/http"
	"time"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/utils"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s, err := cookieStore(filename, filters...)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	return s.ReadCookies(filters...)
}

func (s *epiphanyCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.Open(); err != nil {
		return nil, err
	} else if s.Database == nil {
		return nil, errors.New(`database is nil`)
	}

	var cookies []*kooky.Cookie

	// Epiphany originally used a Mozilla Gecko backend but later switched to WebKit.
	// For possible deviations from the firefox database layout
	// it might be better not to depend on the firefox implementation.

	err := utils.VisitTableRows(s.Database, `moz_cookies`, map[string]string{}, func(rowId *int64, row utils.TableRow) error {
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

		if kooky.FilterCookie(&cookie, filters...) {
			cookies = append(cookies, &cookie)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return cookies, nil
}

// CookieJar returns an initiated http.CookieJar based on the cookies stored by
// the Epiphany/Gnome Web browser. Set cookies are memory stored and do not modify any
// browser files.
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
func CookieStore(filename string, filters ...kooky.Filter) (kooky.CookieStore, error) {
	return cookieStore(filename, filters...)
}

func cookieStore(filename string, filters ...kooky.Filter) (*cookies.CookieJar, error) {
	s := &epiphanyCookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `epiphany`

	return &cookies.CookieJar{CookieStore: s}, nil
}
