//go:build darwin && cgo
// +build darwin,cgo

package chrome

import (
	"errors"
	"fmt"

	keychain "github.com/keybase/go-keychain"
)

// getKeyringPassword retrieves the Chrome Safe Storage password,
// caching it for future calls.
func (s *CookieStore) getKeyringPassword(useSaved bool) ([]byte, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if useSaved && s.KeyringPasswordBytes != nil {
		return s.KeyringPasswordBytes, nil
	}

	kpmKey := `keychain_` + s.BrowserStr
	if useSaved {
		if kpw, ok := keyringPasswordMap.get(kpmKey); ok {
			return kpw, nil
		}
	}

	localSafeStorage := describeSafeStorage(s.BrowserStr)
	password, err := keychain.GetGenericPassword(localSafeStorage.name, localSafeStorage.account, "", "")
	if err != nil {
		return nil, fmt.Errorf(localSafeStorage.errorMsg, localSafeStorage.name, err)
	}
	s.KeyringPasswordBytes = password
	keyringPasswordMap.set(kpmKey, password)

	return s.KeyringPasswordBytes, nil
}

type safeStorage struct {
	name     string
	account  string
	errorMsg string
}

func describeSafeStorage(browserName string) safeStorage {
	defaultStore := safeStorage{
		name:     "Chrome Safe Storage",
		account:  "Chrome",
		errorMsg: "error reading '%s' keychain password: %w",
	}
	if browserName == `edge` {
		defaultStore.name = "Microsoft Edge Safe Storage"
		defaultStore.account = "Microsoft Edge"
	}
	return defaultStore
}
