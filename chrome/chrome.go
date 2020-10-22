package chrome

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-sqlite/sqlite3"

	"github.com/zellyn/kooky"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	var cookies []*kooky.Cookie
	db, err := sqlite3.Open(filename)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	/*
		var version int
		if err := db.VisitTableRecords("meta", func(rowId *int64, rec sqlite3.Record) error {
			if len(rec.Values) != 2 {
				return errors.New(`expected 2 columns for "meta" table`)
			}
			if key, ok := rec.Values[0].(string); ok && key == `version` {
				if vStr, ok := rec.Values[1].(string); ok {
					if v, err := strconv.Atoi(vStr); err == nil {
						version = v
					}
				}
			}
			return nil
		}); err != nil {
			return nil, err
		}
	*/

	var columnIDs = map[string]int{
		// fallback values
		`host_key`:        1, // domain
		`name`:            2,
		`value`:           3,
		`path`:            4,
		`expires_utc`:     5,
		`is_secure`:       6,
		`is_httponly`:     7,
		`encrypted_value`: 12,
	}
	cookiesTableName := `cookies`
	var highestIndex int
	for _, table := range db.Tables() {
		if table.Name() == cookiesTableName {
			for id, column := range table.Columns() {
				name := column.Name()
				if name == `CONSTRAINT` {
					// github.com/go-sqlite/sqlite3.Table.Columns() might report pseudo-columns at the end
					break
				}
				if id > highestIndex {
					highestIndex = id
				}
				columnIDs[name] = id
			}
		}
	}

	err = db.VisitTableRecords(cookiesTableName, func(rowId *int64, rec sqlite3.Record) error {
		if rowId == nil {
			return errors.New(`unexpected nil RowID in Chrome sqlite database`)
		}

		// TODO(zellyn): handle older, shorter rows?
		if lRec := len(rec.Values); lRec < 14 {
			return fmt.Errorf("expected at least 14 columns in cookie file, got: %d", lRec)
		} else if highestIndex > lRec {
			return errors.New(`column index out of bound`)
		}

		cookie := &kooky.Cookie{}

		/*
			-- taken from chrome 80's cookies' sqlite_master
			CREATE TABLE cookies(
				creation_utc INTEGER NOT NULL,
				host_key TEXT NOT NULL,
				name TEXT NOT NULL,
				value TEXT NOT NULL,
				path TEXT NOT NULL,
				expires_utc INTEGER NOT NULL,
				is_secure INTEGER NOT NULL,
				is_httponly INTEGER NOT NULL,
				last_access_utc INTEGER NOT NULL,
				has_expires INTEGER NOT NULL DEFAULT 1,
				is_persistent INTEGER NOT NULL DEFAULT 1,
				priority INTEGER NOT NULL DEFAULT 1,
				encrypted_value BLOB DEFAULT '',
				samesite INTEGER NOT NULL DEFAULT -1,
				source_scheme INTEGER NOT NULL DEFAULT 0,
				UNIQUE (host_key, name, path)
			)
		*/

		domain, ok := rec.Values[columnIDs[`host_key`]].(string)
		if !ok {
			return fmt.Errorf("expected column 2 (host_key) to to be string; got %T", rec.Values[columnIDs[`host_key`]])
		}
		name, ok := rec.Values[columnIDs[`name`]].(string)
		if !ok {
			return fmt.Errorf("expected column 3 (name) in cookie(domain:%s) to to be string; got %T", domain, rec.Values[columnIDs[`name`]])
		}
		value, ok := rec.Values[columnIDs[`value`]].(string)
		if !ok {
			return fmt.Errorf("expected column 4 (value) in cookie(domain:%s, name:%s) to to be string; got %T", domain, name, rec.Values[columnIDs[`value`]])
		}
		path, ok := rec.Values[columnIDs[`path`]].(string)
		if !ok {
			return fmt.Errorf("expected column 5 (path) in cookie(domain:%s, name:%s) to to be string; got %T", domain, name, rec.Values[columnIDs[`path`]])
		}
		var expires_utc int64
		switch i := rec.Values[columnIDs[`expires_utc`]].(type) {
		case int64:
			expires_utc = i
		case int:
			if i != 0 {
				return fmt.Errorf("expected column 6 (expires_utc) in cookie(domain:%s, name:%s) to to be int64 or int with value=0; got %T with value %[3]v", domain, name, rec.Values[columnIDs[`expires_utc`]])
			}
		default:
			return fmt.Errorf("expected column 6 (expires_utc) in cookie(domain:%s, name:%s) to to be int64 or int with value=0; got %T with value %[3]v", domain, name, rec.Values[columnIDs[`expires_utc`]])
		}
		encrypted_value, ok := rec.Values[columnIDs[`encrypted_value`]].([]byte)
		if !ok {
			return fmt.Errorf("expected column 13 (encrypted_value) in cookie(domain:%s, name:%s) to to be []byte; got %T", domain, name, rec.Values[columnIDs[`encrypted_value`]])
		}

		var expiry time.Time
		if expires_utc != 0 {
			expiry = chromeCookieDate(expires_utc)
		}
		creation := chromeCookieDate(*rowId)

		cookie.Domain = domain
		cookie.Name = name
		cookie.Path = path
		cookie.Expires = expiry
		cookie.Creation = creation
		cookie.Secure = rec.Values[columnIDs[`is_secure`]] == 1
		cookie.HttpOnly = rec.Values[columnIDs[`is_httponly`]] == 1

		if len(encrypted_value) > 0 {
			dbFile = filename
			decrypted, err := decryptValue(encrypted_value)
			if err != nil {
				return fmt.Errorf("decrypting cookie %v: %w", cookie, err)
			}
			cookie.Value = decrypted
		} else {
			cookie.Value = value
		}

		if !kooky.FilterCookie(cookie, filters...) {
			return nil
		}

		cookies = append(cookies, cookie)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return cookies, nil
}

// See https://cs.chromium.org/chromium/src/base/time/time.h?l=452&rcl=fceb9a030c182e939a436a540e6dacc70f161cb1
const windowsToUnixMicrosecondsOffset = 116444736e8

// chromeCookieDate converts microseconds to a time.Time object,
// accounting for the switch to Windows epoch (Jan 1 1601).
func chromeCookieDate(timestamp_utc int64) time.Time {
	if timestamp_utc > windowsToUnixMicrosecondsOffset {
		timestamp_utc -= windowsToUnixMicrosecondsOffset
	}

	return time.Unix(timestamp_utc/1e6, (timestamp_utc%1e6)*1e3)
}

var dbFile string // TODO (?)
