package chrome

import (
	"fmt"
	"time"

	"github.com/go-sqlite/sqlite3"
	kooky "github.com/kgoins/kooky/pkg"
)

func ReadChromeCookies(filename string, domainFilter string, nameFilter string, expireAfter time.Time) ([]*kooky.Cookie, error) {
	var cookies []*kooky.Cookie
	db, err := sqlite3.Open(filename)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.VisitTableRecords("cookies", func(rowId *int64, rec sqlite3.Record) error {
		if rowId == nil {
			return fmt.Errorf("unexpected nil RowID in Chrome sqlite database")
		}
		cookie := &kooky.Cookie{}

		// TODO(zellyn): handle older, shorter rows?
		if len(rec.Values) < 14 {
			return fmt.Errorf("expected at least 14 columns in cookie file, got: %d", len(rec.Values))
		}

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

		domain, ok := rec.Values[1].(string)
		if !ok {
			return fmt.Errorf("expected column 2 (host_key) to to be string; got %T", rec.Values[1])
		}
		name, ok := rec.Values[2].(string)
		if !ok {
			return fmt.Errorf("expected column 3 (name) in cookie(domain:%s) to to be string; got %T", domain, rec.Values[2])
		}
		value, ok := rec.Values[3].(string)
		if !ok {
			return fmt.Errorf("expected column 4 (value) in cookie(domain:%s, name:%s) to to be string; got %T", domain, name, rec.Values[3])
		}
		path, ok := rec.Values[4].(string)
		if !ok {
			return fmt.Errorf("expected column 5 (path) in cookie(domain:%s, name:%s) to to be string; got %T", domain, name, rec.Values[4])
		}
		var expires_utc int64
		switch i := rec.Values[5].(type) {
		case int64:
			expires_utc = i
		case int:
			if i != 0 {
				return fmt.Errorf("expected column 6 (expires_utc) in cookie(domain:%s, name:%s) to to be int64 or int with value=0; got %T with value %v", domain, name, rec.Values[5], rec.Values[5])
			}
		default:
			return fmt.Errorf("expected column 6 (expires_utc) in cookie(domain:%s, name:%s) to to be int64 or int with value=0; got %T with value %v", domain, name, rec.Values[5], rec.Values[5])
		}
		encrypted_value, ok := rec.Values[12].([]byte)
		if !ok {
			return fmt.Errorf("expected column 13 (encrypted_value) in cookie(domain:%s, name:%s) to to be []byte; got %T", domain, name, rec.Values[12])
		}

		var expiry time.Time
		if expires_utc != 0 {
			expiry = chromeCookieDate(expires_utc)
		}
		creation := chromeCookieDate(*rowId)

		if domainFilter != "" && domain != domainFilter {
			return nil
		}

		if nameFilter != "" && name != nameFilter {
			return nil
		}

		if !expiry.IsZero() && expiry.Before(expireAfter) {
			return nil
		}

		cookie.Domain = domain
		cookie.Name = name
		cookie.Path = path
		cookie.Expires = expiry
		cookie.Creation = creation
		cookie.Secure = rec.Values[6] == 1
		cookie.HttpOnly = rec.Values[7] == 1

		if len(encrypted_value) > 0 {
			decrypted, err := decryptValue(encrypted_value)
			if err != nil {
				return fmt.Errorf("decrypting cookie %v: %v", cookie, err)
			}
			cookie.Value = decrypted
		} else {
			cookie.Value = value
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
const windowsToUnixMicrosecondsOffset = 11644473600000000

// chromeCookieDate converts microseconds to a time.Time object,
// accounting for the switch to Windows epoch (Jan 1 1601).
func chromeCookieDate(timestamp_utc int64) time.Time {
	if timestamp_utc > windowsToUnixMicrosecondsOffset {
		timestamp_utc -= windowsToUnixMicrosecondsOffset
	}

	return time.Unix(timestamp_utc/1000000, (timestamp_utc%1000000)*1000)
}
