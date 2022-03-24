//go:build !ios && !cgo

package chrome

import (
	"errors"
	"os/exec"
	"strings"
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

	// TODO: use s.browser
	out, err := exec.Command(`/usr/bin/security`, `find-generic-password`, `-s`, `Chrome Safe Storage`, `-wa`, `Chrome`).Output()
	if err != nil {
		return nil, err
	}
	s.KeyringPasswordBytes = []byte(strings.TrimSpace(string(out)))
	keyringPasswordMap.set(kpmKey, s.KeyringPasswordBytes)

	return s.KeyringPasswordBytes, nil
}
