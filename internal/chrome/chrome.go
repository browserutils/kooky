package chrome

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"golang.org/x/crypto/pbkdf2"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/utils"

	"github.com/go-sqlite/sqlite3"
)

// Thanks to https://gist.github.com/dacort/bd6a5116224c594b14db

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

	/*
		var version int
		if err := db.VisitTableRecords("meta", func(rowID *int64, rec sqlite3.Record) error {
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
	for _, table := range s.Database.Tables() {
		if table.Name() == cookiesTableName {
			for id, column := range table.Columns() {
				name := column.Name()
				if _, ok := columnIDs[name]; ok {
					columnIDs[name] = id
				}
			}
		}
	}

	minimumValuesInRow := -1
	for _, index := range columnIDs {
		if index > minimumValuesInRow {
			minimumValuesInRow = index
		}
	}
	minimumValuesInRow++

	err := s.Database.VisitTableRecords(cookiesTableName, func(rowID *int64, rec sqlite3.Record) error {
		if rowID == nil {
			return errors.New(`unexpected nil rowID in Chrome sqlite database`)
		}

		if lRec := len(rec.Values); lRec < minimumValuesInRow {
			return fmt.Errorf("Expected each row to have at least %d values, got %d in row %d", minimumValuesInRow, lRec, rowID)
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

		// https://cs.chromium.org/chromium/src/base/time/time.h?l=452&rcl=fceb9a030c182e939a436a540e6dacc70f161cb1
		var expiry time.Time
		if expires_utc != 0 {
			expiry = utils.FromFILETIME(expires_utc * 10)
		}
		creation := utils.FromFILETIME(*rowID * 10)

		cookie.Domain = domain
		cookie.Name = name
		cookie.Path = path
		cookie.Expires = expiry
		cookie.Creation = creation
		cookie.Secure = rec.Values[columnIDs[`is_secure`]] == 1
		cookie.HttpOnly = rec.Values[columnIDs[`is_httponly`]] == 1

		if len(encrypted_value) > 0 {
			decrypted, err := s.decrypt(encrypted_value)
			if err != nil {
				return fmt.Errorf("decrypting cookie %v: %w", cookie, err)
			}
			cookie.Value = string(decrypted)
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

// "mock_password" from https://github.com/chromium/chromium/blob/34f6b421d6d255b27e01d82c3c19f49a455caa06/crypto/mock_apple_keychain.cc#L75
var (
	fallbackPasswordLinux = [...]byte{'p', 'e', 'a', 'n', 'u', 't', 's'}
	fallbackPasswordMacOS = [...]byte{'m', 'o', 'c', 'k', '_', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd'}                     // mock keychain
	prefixDPAPI           = [...]byte{1, 0, 0, 0, 208, 140, 157, 223, 1, 21, 209, 17, 140, 122, 0, 192, 79, 194, 151, 235} // 0x01000000D08C9DDF0115D1118C7A00C04FC297EB
)

// key might be the absolute path of the `Local State` file containing the encrypted key
// or a similar identifier
var keyringPasswordMap = keyringPasswordMapType{
	v: make(map[string][]byte),
}

type keyringPasswordMapType struct {
	mu sync.RWMutex
	v  map[string][]byte
}

func (k *keyringPasswordMapType) get(key string) (val []byte, ok bool) {
	if k == nil {
		return
	}
	k.mu.RLock()
	defer k.mu.RUnlock()
	val, ok = k.v[key]
	return val, ok
}
func (k *keyringPasswordMapType) set(key string, val []byte) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.v[key] = val
}

func (s *CookieStore) decrypt(encrypted []byte) ([]byte, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if len(encrypted) == 0 {
		return nil, errors.New(`empty encrypted value`)
	}

	if len(encrypted) <= 3 {
		return nil, fmt.Errorf(`encrypted value is too short (%d<=3)`, len(encrypted))
	}

	// try to reuse previously successful decryption method
	if s.DecryptionMethod != nil {
		decrypted, err := s.DecryptionMethod(encrypted, s.PasswordBytes)
		if err == nil {
			return decrypted, nil
		} else {
			s.DecryptionMethod = nil
		}
	}

	var decrypt func(encrypted, password []byte) ([]byte, error)

	// prioritize previously selected platform then current platform and then other platforms in order of usage on non-server computers
	// TODO: mobile
	var osMap = map[string]struct{}{} // used for deduplication
	var oss []string
	for _, opsys := range []string{s.OSStr, runtime.GOOS, `windows`, `darwin`, `linux`} {
		if _, ok := osMap[opsys]; ok {
			continue
		}
		oss = append(oss, opsys)
		osMap[opsys] = struct{}{}
	}

	for _, opsys := range oss {
		// "useSavedKeyringPassword" and "tryNr" have to preserve state between retries
		var useSavedKeyringPassword bool = true
		var tryNr int
	tryAgain:
		var password, keyringPassword, fallbackPassword []byte
		var needsKeyringQuerying bool
		switch opsys {
		case `windows`:
			switch {
			case bytes.HasPrefix(encrypted, prefixDPAPI[:]):
				// present before Chrome v80 on Windows
				decrypt = func(encrypted, _ []byte) ([]byte, error) {
					return decryptDPAPI(encrypted)
				}
			case bytes.HasPrefix(encrypted, []byte(`v10`)):
				fallthrough
			default:
				needsKeyringQuerying = true
				decrypt = decryptAES256GCM
			}
		case `darwin`:
			needsKeyringQuerying = true
			fallbackPassword = fallbackPasswordMacOS[:]
			decrypt = func(encrypted, password []byte) ([]byte, error) {
				return decryptAESCBC(encrypted, password, aescbcIterationsMacOS)
			}
		case `linux`:
			switch {
			case bytes.HasPrefix(encrypted, []byte(`v11`)):
				needsKeyringQuerying = true
				fallbackPassword = fallbackPasswordLinux[:]
			case bytes.HasPrefix(encrypted, []byte(`v10`)):
				password = fallbackPasswordLinux[:]
			default:
				password = fallbackPasswordLinux[:]
			}
			decrypt = func(encrypted, password []byte) ([]byte, error) {
				return decryptAESCBC(encrypted, password, aescbcIterationsLinux)
			}
		}
		if decrypt == nil {
			continue
		}

		if needsKeyringQuerying {
			switch tryNr {
			case 0, 1:
				pw, err := s.getKeyringPassword(useSavedKeyringPassword)
				if err == nil {
					password = pw
				} else {
					password = fallbackPassword
					tryNr = 2 // skip querying
				}
				// query keyring passwords on try #1 without simply returning saved ones
				useSavedKeyringPassword = false
			case 2:
				password = fallbackPassword
			}
			tryNr++
		}

		decrypted, err := decrypt(encrypted, password)
		if err == nil {
			s.DecryptionMethod = decrypt
			s.OSStr = opsys
			s.PasswordBytes = password
			if len(keyringPassword) > 0 {
				s.KeyringPasswordBytes = keyringPassword
			}
			return decrypted, nil
		} else if tryNr > 0 && tryNr < 3 {
			goto tryAgain
		}
	}

	return nil, errors.New(`unknown encryption method`)
}

const (
	aescbcSalt            = `saltysalt`
	aescbcIV              = `                `
	aescbcIterationsLinux = 1
	aescbcIterationsMacOS = 1003
	aescbcLength          = 16
)

func decryptAESCBC(encrypted, password []byte, iterations int) ([]byte, error) {
	if len(encrypted) == 0 {
		return nil, errors.New("empty encrypted value")
	}

	if len(encrypted) <= 3 {
		return nil, fmt.Errorf("too short encrypted value (%d<=3)", len(encrypted))
	}

	// strip "v##"
	encrypted = encrypted[3:]

	key := pbkdf2.Key(password, []byte(aescbcSalt), iterations, aescbcLength, sha1.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	decrypted := make([]byte, len(encrypted))
	cbc := cipher.NewCBCDecrypter(block, []byte(aescbcIV))
	cbc.CryptBlocks(decrypted, encrypted)

	// In the padding scheme the last <padding length> bytes
	// have a value equal to the padding length, always in (1,16]
	if len(decrypted)%aescbcLength != 0 {
		return nil, fmt.Errorf("decrypted data block length is not a multiple of %d", aescbcLength)
	}
	paddingLen := int(decrypted[len(decrypted)-1])
	if paddingLen > 16 {
		return nil, fmt.Errorf("invalid last block padding length: %d", paddingLen)
	}

	return decrypted[:len(decrypted)-paddingLen], nil
}

func decryptAES256GCM(encrypted, password []byte) ([]byte, error) {
	// https://stackoverflow.com/a/60423699

	if len(encrypted) < 3+12+16 {
		return nil, errors.New(`encrypted value too short`)
	}

	/* encoded value consists of: {
		"v10" (3 bytes)
		nonce (12 bytes)
		ciphertext (variable size)
		tag (16 bytes)
	}
	*/
	nonce := encrypted[3 : 3+12]
	ciphertextWithTag := encrypted[3+12:]

	block, err := aes.NewCipher(password)
	if err != nil {
		return nil, err
	}

	// default size for nonce and tag match
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plaintext, err := aesgcm.Open(nil, nonce, ciphertextWithTag, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
