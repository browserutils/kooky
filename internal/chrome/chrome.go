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

	"golang.org/x/crypto/pbkdf2"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/timex"
	"github.com/zellyn/kooky/internal/utils"
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

	headerMappings := map[string]string{
		"secure":   "is_secure",
		"httponly": "is_httponly",
	}
	err := utils.VisitTableRows(s.Database, `cookies`, headerMappings, func(rowID *int64, row utils.TableRow) error {
		cookie := &kooky.Cookie{
			Creation: timex.FromFILETIME(*rowID * 10),
		}

		var err error

		cookie.Domain, err = row.String(`host_key`)
		if err != nil {
			return err
		}

		cookie.Name, err = row.String(`name`)
		if err != nil {
			return err
		}

		cookie.Path, err = row.String(`path`)
		if err != nil {
			return err
		}

		if expires_utc, err := row.Int64(`expires_utc`); err == nil {
			// https://cs.chromium.org/chromium/src/base/time/time.h?l=452&rcl=fceb9a030c182e939a436a540e6dacc70f161cb1
			if expires_utc != 0 {
				cookie.Expires = timex.FromFILETIME(expires_utc * 10)
			}
		} else {
			return err
		}

		cookie.Secure, err = row.Bool(`is_secure`)
		if err != nil {
			return err
		}

		cookie.HttpOnly, err = row.Bool(`is_httponly`)
		if err != nil {
			return err
		}

		encrypted_value, err := row.BytesStringOrFallback(`encrypted_value`, nil)
		if err != nil {
			return err
		} else if len(encrypted_value) > 0 {
			if decrypted, err := s.decrypt(encrypted_value); err == nil {
				cookie.Value = string(decrypted)
			} else {
				return fmt.Errorf("decrypting cookie %v: %w", cookie, err)
			}
		} else {
			cookie.Value, err = row.String(`value`)
			if err != nil {
				return err
			}
		}

		if kooky.FilterCookie(cookie, filters...) {
			cookies = append(cookies, cookie)
		}

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
