// +build !darwin,!windows,!linux android

package chrome

import "errors"

func (s *CookieStore) getKeyringPassword(useSaved) ([]byte, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if useSaved && s.KeyringPassword != nil {
		return s.KeyringPassword, nil
	}

	return nil, errors.New(`not implemented`)
}
