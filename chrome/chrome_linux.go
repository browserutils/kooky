// +build !android

package chrome

// https://cs.chromium.org/chromium/src/components/os_crypt/os_crypt_linux.cc?q=peanuts     // password "peanuts"   for v10
// https://cs.chromium.org/chromium/src/components/os_crypt/os_crypt_linux.cc?q=saltysalt   // salt     "saltysalt"

// https://redd.it/39swuj/
// https://n8henrie.com/2014/05/decrypt-chrome-cookies-with-python/
// https://github.com/obsidianforensics/hindsight/blob/311c80ff35b735b273d69529d3e024d1b1aa2796/pyhindsight/browsers/chrome.py#L432

// https://gist.github.com/dacort/bd6a5116224c594b14db

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"errors"
	"fmt"

	"golang.org/x/crypto/pbkdf2"

	secret_service "github.com/zalando/go-keyring/secret_service"
)

const (
	salt       = "saltysalt"
	iv         = "                "
	length     = 16
	iterations = 1
)

// keychainPassword is a cache of the password read from the keychain.
var keychainPassword []byte

// setChromeKeychainPassword exists so tests can avoid trying to read
// the Keychain.
func setChromeKeychainPassword(password []byte) []byte {
	oldPassword := keychainPassword
	keychainPassword = password
	return oldPassword
}

func queryDbus(browser string) ([]byte, error) {
	// this is mostly a copy from github.com/zalando/go-keyring (MIT License)
	// Get()      from https://github.com/zalando/go-keyring/blob/07372e614fb45baa337eaca014ed232b7b196200/keyring_linux.go#L77
	// findItem() from https://github.com/zalando/go-keyring/blob/07372e614fb45baa337eaca014ed232b7b196200/keyring_linux.go#L51

	type secretServiceProvider struct{}

	svc, err := secret_service.NewSecretService()
	if err != nil {
		return []byte{}, err
	}

	collection := svc.GetLoginCollection()

	search := map[string]string{
		"application": browser,
	}

	if err := svc.Unlock(collection.Path()); err != nil {
		return []byte{}, err
	}

	results, err := svc.SearchItems(collection, search)
	if err != nil {
		return []byte{}, err
	}
	if len(results) == 0 {
		return []byte{}, fmt.Errorf("secret not found in keyring")
	}
	item := results[0]

	// open a session
	session, err := svc.OpenSession()
	if err != nil {
		return []byte{}, err
	}
	defer svc.Close(session)

	secret, err := svc.GetSecret(item, session.Path())
	if err != nil {
		return []byte{}, err
	}

	// secret.Value is base64 standard encoded
	return secret.Value, nil
}

// getKeychainPassword retrieves the Chrome Safe Storage password,
// caching it for future calls.
func getKeychainPassword() ([]byte, error) {
	// https://cs.chromium.org/chromium/src/components/os_crypt/key_storage_linux.cc?q="chromium+safe+storage"

	// v10 cookies
	defaultPassword := []byte("peanuts")
	password := defaultPassword

	// v11 cookies  - chromium --password-store=gnome
	if pw, err := queryDbus("chrome"); err == nil && len(pw) > 0 {
		password = pw
	} else if pw, err := queryDbus("chromium"); err == nil && len(pw) > 0 {
		password = pw
	}

	return password, nil
}

func decryptValue(encrypted []byte) (string, error) {
	if len(encrypted) == 0 {
		return "", errors.New("empty encrypted value")
	}

	if len(encrypted) <= 3 {
		return "", fmt.Errorf("too short encrypted value (%d<=3)", len(encrypted))
	}

	var password []byte
	switch string(encrypted[:3]) {
	case `v10`:
		password = []byte(`peanuts`)
	case `v11`:
		pw, err := getKeychainPassword()
		if err != nil {
			return "", err
		}
		password = pw
	default:
		password = []byte(`peanuts`)
	}

	encrypted = encrypted[3:]

	key := pbkdf2.Key(password, []byte(salt), iterations, length, sha1.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	decrypted := make([]byte, len(encrypted))
	cbc := cipher.NewCBCDecrypter(block, []byte(iv))
	cbc.CryptBlocks(decrypted, encrypted)

	plainText, err := aesStripPadding(decrypted)
	if err != nil {
		return "", err
	}
	return string(plainText), nil
}

// In the padding scheme the last <padding length> bytes
// have a value equal to the padding length, always in (1,16]
func aesStripPadding(data []byte) ([]byte, error) {
	if len(data)%length != 0 {
		return nil, fmt.Errorf("decrypted data block length is not a multiple of %d", length)
	}
	paddingLen := int(data[len(data)-1])
	if paddingLen > 16 {
		return nil, fmt.Errorf("invalid last block padding length: %d", paddingLen)
	}
	return data[:len(data)-paddingLen], nil
}
