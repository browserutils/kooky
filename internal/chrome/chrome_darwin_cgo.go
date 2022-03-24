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

	password, err := keychain.GetGenericPassword("Chrome Safe Storage", "Chrome", "", "")
	if err != nil {
		return nil, fmt.Errorf("error reading 'Chrome Safe Storage' keychain password: %w", err)
	}
	s.KeyringPasswordBytes = password
	keyringPasswordMap.set(kpmKey, password)

	return s.KeyringPasswordBytes, nil
}
