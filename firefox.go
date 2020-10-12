package kooky

import (
	"fmt"
	"time"

	"github.com/bobesa/go-domain-util/domainutil"
	"github.com/go-sqlite/sqlite3"
)

func ReadFirefoxCookies(filename string) ([]*Cookie, error) {
	var cookies []*Cookie
	db, err := sqlite3.Open(filename)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.VisitTableRecords("moz_cookies", func(rowId *int64, rec sqlite3.Record) error {
		if lRec := len(rec.Values); lRec != 13 && lRec != 14 {
			return fmt.Errorf("got %d columns, but expected 13 or 14", lRec)
		}

		cookie := Cookie{}
		var ok bool
		var columnShift int

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

		switch rec.Values[6].(type) {
		case int32, uint64:
			columnShift = -1 // "baseDomain" column was removed
		}

		// Name
		cookie.Name, ok = rec.Values[3+columnShift].(string)
		if !ok {
			return fmt.Errorf("got unexpected value for Name %v", rec.Values[3+columnShift])
		}

		// Value
		cookie.Value, ok = rec.Values[4+columnShift].(string)
		if !ok {
			return fmt.Errorf("got unexpected value for Value %v", rec.Values[4+columnShift])
		}

		// Domain
		if columnShift == 0 {
			cookie.Domain, ok = rec.Values[1].(string)
			if !ok {
				return fmt.Errorf("got unexpected value for Domain %v", rec.Values[1])
			}
		} else {
			if host, ok := rec.Values[4].(string); ok {
				cookie.Domain = domainutil.Domain(host)
			} else {
				return fmt.Errorf("got unexpected value for Host %v", rec.Values[4])
			}

		}

		// Path
		cookie.Path, ok = rec.Values[6+columnShift].(string)
		if !ok {
			return fmt.Errorf("got unexpected value for Path %v", rec.Values[6+columnShift])
		}

		// Expires
		if int32Value, ok := rec.Values[7+columnShift].(int32); ok {
			cookie.Expires = time.Unix(int64(int32Value), 0)
		} else if uint64Value, ok := rec.Values[7+columnShift].(uint64); ok {
			cookie.Expires = time.Unix(int64(uint64Value), 0)
		} else {
			return fmt.Errorf("got unexpected value for Expires %v (type %T)", rec.Values[7+columnShift], rec.Values[7+columnShift])
		}

		// Creation
		int64Value, ok := rec.Values[9+columnShift].(int64)
		if !ok {
			return fmt.Errorf("got unexpected value for Creation %v (type %T)", rec.Values[9+columnShift], rec.Values[9+columnShift])
		}
		cookie.Creation = time.Unix(int64Value/1e6, 0) // drop nanoseconds

		// Secure
		intValue, ok := rec.Values[10+columnShift].(int)
		if !ok {
			return fmt.Errorf("got unexpected value for Secure %v", rec.Values[10+columnShift])
		}
		cookie.Secure = intValue > 0

		// HttpOnly
		intValue, ok = rec.Values[11+columnShift].(int)
		if !ok {
			return fmt.Errorf("got unexpected value for HttpOnly %v", rec.Values[11+columnShift])
		}
		cookie.HttpOnly = intValue > 0

		cookies = append(cookies, &cookie)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return cookies, nil
}
