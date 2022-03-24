//go:build linux && !android

package chrome

import (
	"errors"
	"fmt"

	secret_service "github.com/zalando/go-keyring/secret_service"
)

// https://cs.chromium.org/chromium/src/components/os_crypt/os_crypt_linux.cc?q=peanuts     // password "peanuts"   for v10
// https://cs.chromium.org/chromium/src/components/os_crypt/os_crypt_linux.cc?q=saltysalt   // salt     "saltysalt"

// https://redd.it/39swuj/
// https://n8henrie.com/2014/05/decrypt-chrome-cookies-with-python/
// https://github.com/obsidianforensics/hindsight/blob/311c80ff35b735b273d69529d3e024d1b1aa2796/pyhindsight/browsers/chrome.py#L432

// https://gist.github.com/dacort/bd6a5116224c594b14db

// getKeyringPassword retrieves the Chrome Safe Storage password,
// caching it for future calls.
func (s *CookieStore) getKeyringPassword(useSaved bool) ([]byte, error) {
	// https://cs.chromium.org/chromium/src/components/os_crypt/key_storage_linux.cc?q="chromium+safe+storage"

	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if useSaved && s.KeyringPasswordBytes != nil {
		return s.KeyringPasswordBytes, nil
	}

	// chromium --password-store=gnome

	// this is mostly a copy from github.com/zalando/go-keyring (MIT License)
	// Get()      from https://github.com/zalando/go-keyring/blob/07372e614fb45baa337eaca014ed232b7b196200/keyring_linux.go#L77
	// findItem() from https://github.com/zalando/go-keyring/blob/07372e614fb45baa337eaca014ed232b7b196200/keyring_linux.go#L51

	browser := s.BrowserStr
	switch browser {
	case `chrome`:
	case `chromium`:
	case `brave`:
	case `opera`:
		browser = `chromium`
	default:
		// TODO:
		return nil, errors.New(`unknown browser`)
	}

	kpmKey := `dbus_` + browser
	if useSaved {
		if kpw, ok := keyringPasswordMap.get(kpmKey); ok {
			return kpw, nil
		}
	}

	type secretServiceProvider struct{}

	svc, err := secret_service.NewSecretService()
	if err != nil {
		return nil, err
	}

	collection := svc.GetLoginCollection()
	if err := svc.Unlock(collection.Path()); err != nil {
		return nil, err
	}

	search := map[string]string{
		"application": browser,
	}
	results, err := svc.SearchItems(collection, search)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("secret not found in keyring")
	}
	item := results[0]

	// open a session
	session, err := svc.OpenSession()
	if err != nil {
		return nil, err
	}
	defer svc.Close(session)

	secret, err := svc.GetSecret(item, session.Path())
	if err != nil {
		return nil, err
	}

	s.KeyringPasswordBytes = secret.Value
	keyringPasswordMap.set(kpmKey, secret.Value)

	// s.KeyringPassword is base64 standard encoded - do not decode!
	return s.KeyringPasswordBytes, nil
}
