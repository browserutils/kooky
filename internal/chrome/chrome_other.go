//go:build (!darwin && !windows && !linux) || android || ios

package chrome

import "errors"

func (s *CookieStore) getKeyringPassword(useSaved bool) ([]byte, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if useSaved && s.KeyringPasswordBytes != nil {
		return s.KeyringPasswordBytes, nil
	}

	return nil, errors.New(`not implemented`)
}
