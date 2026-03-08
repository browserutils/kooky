//go:build linux && !android

package chrome

import (
	"errors"
	"fmt"
	"os"

	"github.com/godbus/dbus/v5"
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

	// KDE uses the native KWallet D-Bus API for password storage;
	// the freedesktop Secret Service search does not find keys stored there.
	// https://chromium.googlesource.com/chromium/src/+/master/docs/linux/password_storage.md
	var pw []byte
	var err error
	kdeVer, isKDE := os.LookupEnv(`KDE_SESSION_VERSION`)
	if isKDE {
		pw, err = s.getKWalletPassword(kdeVer)
		if err != nil {
			pw, err = s.getSecretServicePassword(browser)
		}
	} else {
		pw, err = s.getSecretServicePassword(browser)
		if err != nil {
			pw, err = s.getKWalletPassword(``)
		}
	}
	if err != nil {
		return nil, err
	}

	s.KeyringPasswordBytes = pw
	keyringPasswordMap.set(kpmKey, pw)

	// password is base64 standard encoded - do not decode!
	return s.KeyringPasswordBytes, nil
}

func (s *CookieStore) getSecretServicePassword(browser string) ([]byte, error) {
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

	session, err := svc.OpenSession()
	if err != nil {
		return nil, err
	}
	defer svc.Close(session)

	secret, err := svc.GetSecret(item, session.Path())
	if err != nil {
		return nil, err
	}

	return secret.Value, nil
}

// getKWalletPassword retrieves the safe storage password via the KWallet D-Bus API.
// Chromium stores the key under folder "<Account> Keys",
// entry "<Account> Safe Storage" (e.g. "Chromium Keys"/"Chromium Safe Storage").
// https://chromium.googlesource.com/chromium/src/+/master/docs/linux/password_storage.md
func (s *CookieStore) getKWalletPassword(kdeVer string) ([]byte, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, fmt.Errorf("kwallet: session bus: %w", err)
	}

	account := s.safeStorageAccount() // e.g. "Chromium", "Chrome"
	entry := s.safeStorageName()      // e.g. "Chromium Safe Storage"
	folder := account + ` Keys`       // e.g. "Chromium Keys"
	appID := `kooky`

	// try the matching KDE version first, then others
	suffixes := []string{`6`, `5`, ``}
	if len(kdeVer) > 0 {
		prioritized := []string{kdeVer}
		for _, s := range suffixes {
			if s != kdeVer {
				prioritized = append(prioritized, s)
			}
		}
		suffixes = prioritized
	}
	for _, suffix := range suffixes {
		svcName := `org.kde.kwalletd` + suffix
		objPath := dbus.ObjectPath(`/modules/kwalletd` + suffix)
		obj := conn.Object(svcName, objPath)

		var walletName string
		if err := obj.Call(`org.kde.KWallet.networkWallet`, 0).Store(&walletName); err != nil {
			continue
		}

		var handle int32
		if err := obj.Call(`org.kde.KWallet.open`, 0, walletName, int64(0), appID).Store(&handle); err != nil {
			continue
		}
		if handle < 0 {
			continue
		}

		var pw string
		if err := obj.Call(`org.kde.KWallet.readPassword`, 0, handle, folder, entry, appID).Store(&pw); err != nil {
			continue
		}
		if len(pw) == 0 {
			continue
		}

		return []byte(pw), nil
	}

	return nil, errors.New(`kwallet: password not found`)
}
