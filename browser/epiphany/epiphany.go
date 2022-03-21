package epiphany

import (
	"errors"
	"fmt"
	"time"

	"github.com/zellyn/kooky"

	"github.com/go-sqlite/sqlite3"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s := &epiphanyCookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `epiphany`

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

	var columnIDs = map[string]int{
		// fallback values
		`name`:       1,
		`value`:      2,
		`host`:       3,
		`path`:       4,
		`expiry`:     5,
		`isSecure`:   7,
		`isHttpOnly`: 8,
	}
	cookiesTableName := `moz_cookies`
	var highestIndex int
	for _, table := range s.Database.Tables() {
		if table.Name() == cookiesTableName {
			for id, column := range table.Columns() {
				name := column.Name()
				if name == `CONSTRAINT` {
					// github.com/go-sqlite/sqlite3.Table.Columns() reports pseudo-columns for host, path, originAttributes
					break
				}
				if id > highestIndex {
					highestIndex = id
				}
				columnIDs[name] = id
			}
		}
	}

	err := s.Database.VisitTableRecords(cookiesTableName, func(rowId *int64, rec sqlite3.Record) error {
		if lRec := len(rec.Values); lRec != 9 {
			return fmt.Errorf("got %d columns, but expected 9", lRec)
		} else if highestIndex > lRec {
			return errors.New(`column index out of bound`)
		}

		cookie := kooky.Cookie{}
		var ok bool

		/*
			-- Epiphany 3.32 - copied from sqlitebrowser
			CREATE TABLE moz_cookies (
				id INTEGER PRIMARY KEY,
				name TEXT,
				value TEXT
				host TEXT,
				path TEXT,
				expiry INTEGER,
				lastAccessed INTEGER,
				isSecure INTEGER,
				isHttpOnly INTEGER
			)
		*/

		// Name
		cookie.Name, ok = rec.Values[columnIDs[`name`]].(string)
		if !ok {
			return fmt.Errorf("got unexpected value for Name %v (type %[1]T)", rec.Values[columnIDs[`name`]])
		}

		// Value
		cookie.Value, ok = rec.Values[columnIDs[`value`]].(string)
		if !ok {
			return fmt.Errorf("got unexpected value for Value %v (type %[1]T)", rec.Values[columnIDs[`value`]])
		}

		// Host
		cookie.Domain, ok = rec.Values[columnIDs[`host`]].(string)
		if !ok {
			return fmt.Errorf("got unexpected value for Host %v (type %[1]T)", rec.Values[columnIDs[`host`]])
		}

		// Path
		cookie.Path, ok = rec.Values[columnIDs[`path`]].(string)
		if !ok {
			return fmt.Errorf("got unexpected value for Path %v (type %[1]T)", rec.Values[columnIDs[`path`]])
		}

		// Expires
		{
			expiry := rec.Values[columnIDs[`expiry`]]
			if int32Value, ok := expiry.(int32); ok {
				cookie.Expires = time.Unix(int64(int32Value), 0)
			} else if uint64Value, ok := expiry.(uint64); ok {
				cookie.Expires = time.Unix(int64(uint64Value), 0)
			} else {
				return fmt.Errorf("got unexpected value for Expires %v (type %[1]T)", expiry)
			}
		}

		// Secure
		intValue, ok := rec.Values[columnIDs[`isSecure`]].(int)
		if !ok {
			return fmt.Errorf("got unexpected value for Secure %v (type %[1]T)", rec.Values[columnIDs[`isSecure`]])
		}
		cookie.Secure = intValue > 0

		// HttpOnly
		intValue, ok = rec.Values[columnIDs[`isHttpOnly`]].(int)
		if !ok {
			return fmt.Errorf("got unexpected value for HttpOnly %v (type %[1]T)", rec.Values[columnIDs[`isHttpOnly`]])
		}
		cookie.HttpOnly = intValue > 0

		if !kooky.FilterCookie(&cookie, filters...) {
			return nil
		}

		cookies = append(cookies, &cookie)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return cookies, nil
}
