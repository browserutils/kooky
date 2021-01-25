package firefox

import (
	"errors"
	"fmt"
	"time"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/utils"

	"github.com/bobesa/go-domain-util/domainutil"
)

func (s *CookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.Open(); err != nil {
		return nil, err
	} else if s.Database == nil {
		return nil, errors.New(`database is nil`)
	}

	var cookies []*kooky.Cookie

	err := utils.VisitTableRows(s.Database, `moz_cookies`, map[string]string{}, func(rowId *int64, row utils.TableRow) error {
		/*
			-- Firefox 78 ESR - copied from sqlitebrowser
			CREATE TABLE moz_cookies(
				id INTEGER PRIMARY KEY,
				originAttributes TEXT NOT NULL DEFAULT '',
				name TEXT,
				value TEXT,
				host TEXT,
				path TEXT,
				expiry INTEGER,
				lastAccessed INTEGER,
				creationTime INTEGER,
				isSecure INTEGER,
				isHttpOnly INTEGER,
				inBrowserElement INTEGER DEFAULT 0,
				sameSite INTEGER DEFAULT 0,
				rawSameSite INTEGER DEFAULT 0,
				CONSTRAINT moz_uniqueid UNIQUE (name, host, path, originAttributes)
			)
		*/

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

		// Domain
		if baseDomain := row.ValueOrFallback(`baseDomain`, nil); baseDomain == nil {
			if host, err := row.String(`host`); err != nil {
				return err
			} else {
				cookie.Domain = domainutil.Domain(host)
			}
		} else {
			// handle databases prior v78 ESR
			var ok bool
			cookie.Domain, ok = baseDomain.(string)
			if !ok {
				return fmt.Errorf("got unexpected value for baseDomain %v (type %[1]T)", baseDomain)
			}
		}

		// Path
		cookie.Path, err = row.String(`path`)
		if err != nil {
			return err
		}

		// Expires
		if expiry, err := row.Int64(`expiry`); err == nil {
			cookie.Expires = time.Unix(expiry, 0)
		} else {
			return err
		}

		// Creation
		if creationTime, err := row.Int64(`creationTime`); err == nil {
			cookie.Creation = time.Unix(creationTime/1e6, 0) // drop nanoseconds
		} else {
			return err
		}

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
