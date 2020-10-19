package chrome

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"errors"
	"fmt"

	"golang.org/x/crypto/pbkdf2"

	keychain "github.com/keybase/go-keychain"
)

// Thanks to https://gist.github.com/dacort/bd6a5116224c594b14db.

const (
	salt       = "saltysalt"
	iv         = "                "
	length     = 16
	iterations = 1003
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

// getKeychainPassword retrieves the Chrome Safe Storage password,
// caching it for future calls.
func getKeychainPassword() ([]byte, error) {
	if keychainPassword == nil {
		password, err := keychain.GetGenericPassword("Chrome Safe Storage", "Chrome", "", "")
		if err != nil {
			return nil, fmt.Errorf("error reading 'Chrome Safe Storage' keychain password: %v", err)
		}
		keychainPassword = password
	}
	return keychainPassword, nil
}

func decryptValue(encrypted []byte) (string, error) {
	if len(encrypted) == 0 {
		return "", errors.New("empty encrypted value")
	}

	if len(encrypted) <= 3 {
		return "", fmt.Errorf("too short encrypted value (%d<=3)", len(encrypted))
	}

	encrypted = encrypted[3:]

	password, err := getKeychainPassword()
	if err != nil {
		return "", err
	}

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
