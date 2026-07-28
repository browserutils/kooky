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

const (
	secretServiceName = `org.freedesktop.secrets`
	secretServicePath = `/org/freedesktop/secrets`
	// defaultCollectionAlias is where gnome-keyring points the "default" alias.
	// Chrome Safe Storage usually lives here ("Default keyring"), not in "login".
	defaultCollectionAlias = `/org/freedesktop/secrets/aliases/default`
	loginCollectionPath    = `/org/freedesktop/secrets/collection/login`
)

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

	browser := s.safeStorageApplication()

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
	var primaryErr, fallbackErr error
	kdeVer, isKDE := os.LookupEnv(`KDE_SESSION_VERSION`)
	if isKDE {
		pw, primaryErr = s.getKWalletPassword(kdeVer)
		if primaryErr != nil {
			pw, fallbackErr = s.getSecretServicePassword(browser)
		}
	} else {
		pw, primaryErr = s.getSecretServicePassword(browser)
		if primaryErr != nil {
			pw, fallbackErr = s.getKWalletPassword(``)
		}
	}
	if len(pw) == 0 {
		switch {
		case primaryErr != nil && fallbackErr != nil:
			return nil, fmt.Errorf("%w; fallback: %v", primaryErr, fallbackErr)
		case primaryErr != nil:
			return nil, primaryErr
		case fallbackErr != nil:
			return nil, fallbackErr
		default:
			return nil, errors.New(`keyring password not found`)
		}
	}

	s.KeyringPasswordBytes = pw
	keyringPasswordMap.set(kpmKey, pw)

	// password is base64 standard encoded - do not decode!
	return s.KeyringPasswordBytes, nil
}

func (s *CookieStore) getSecretServicePassword(browser string) ([]byte, error) {
	// chromium --password-store=gnome
	//
	// Historically this only searched go-keyring's "login" collection. That
	// misses Chrome Safe Storage on common gnome-keyring setups where the
	// secret lives in the default keyring (aliases/default → Default_5fkeyring)
	// while a separate empty/unrelated "login" collection still exists.
	// secret-tool uses Service.SearchItems (all collections); we do the same,
	// then fall back to per-collection search for older implementations.

	svc, err := secret_service.NewSecretService()
	if err != nil {
		return nil, err
	}

	search := map[string]string{
		"application": browser,
	}

	if pw, err := secretServicePasswordFromServiceSearch(svc, search); err == nil && len(pw) > 0 {
		return pw, nil
	}

	var lastErr error
	for _, collection := range secretServiceCollections(svc) {
		// Best-effort unlock. go-keyring's Unlock is strict about path identity
		// (alias vs resolved path) and errors when a collection is already
		// unlocked; SearchItems/GetSecret still work on unlocked collections.
		_ = svc.Unlock(collection.Path())

		results, err := svc.SearchItems(collection, search)
		if err != nil {
			lastErr = err
			continue
		}
		if len(results) == 0 {
			continue
		}

		pw, err := secretServiceReadItem(svc, results[0])
		if err != nil {
			lastErr = err
			continue
		}
		if len(pw) > 0 {
			return pw, nil
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("secret not found in keyring for application %q", browser)
}

// secretServicePasswordFromServiceSearch uses org.freedesktop.Secret.Service.SearchItems,
// which searches every collection (same as secret-tool).
func secretServicePasswordFromServiceSearch(svc *secret_service.SecretService, search map[string]string) ([]byte, error) {
	obj := svc.Object(secretServiceName, secretServicePath)

	var unlocked, locked []dbus.ObjectPath
	if err := obj.Call(`org.freedesktop.Secret.Service.SearchItems`, 0, search).Store(&unlocked, &locked); err != nil {
		return nil, err
	}

	if len(locked) > 0 {
		// Best-effort unlock of locked matches; ignore prompt failures so we
		// can still try already-unlocked results.
		var unlockedPaths []dbus.ObjectPath
		var prompt dbus.ObjectPath
		_ = obj.Call(`org.freedesktop.Secret.Service.Unlock`, 0, locked).Store(&unlockedPaths, &prompt)
	}

	candidates := make([]dbus.ObjectPath, 0, len(unlocked)+len(locked))
	candidates = append(candidates, unlocked...)
	candidates = append(candidates, locked...)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("secret not found in keyring")
	}

	for _, item := range candidates {
		pw, err := secretServiceReadItem(svc, item)
		if err != nil || len(pw) == 0 {
			continue
		}
		return pw, nil
	}
	return nil, fmt.Errorf("secret not found in keyring")
}

func secretServiceReadItem(svc *secret_service.SecretService, item dbus.ObjectPath) ([]byte, error) {
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

// secretServiceCollections returns Secret Service collections to search, in
// priority order: default alias (Chrome's usual home), login, then any others.
func secretServiceCollections(svc *secret_service.SecretService) []dbus.BusObject {
	preferred := []dbus.ObjectPath{
		defaultCollectionAlias,
		loginCollectionPath,
	}

	var out []dbus.BusObject
	seen := map[string]struct{}{}
	add := func(path dbus.ObjectPath) {
		if path == `` || path == `/` {
			return
		}
		key := string(path)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, svc.Object(secretServiceName, path))
	}

	for _, p := range preferred {
		add(p)
	}

	obj := svc.Object(secretServiceName, secretServicePath)
	if val, err := obj.GetProperty(`org.freedesktop.Secret.Service.Collections`); err == nil {
		if paths, ok := val.Value().([]dbus.ObjectPath); ok {
			for _, p := range paths {
				add(p)
			}
		}
	}
	return out
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
		if err := obj.Call(`org.kde.KWallet.readPassword`, 0, handle, folder, entry, appID).Store(&pw); err == nil && len(pw) > 0 {
			return []byte(pw), nil
		}

		// Fallback: try xdg-desktop-portal binary entry (used by Vivaldi, newer Chromium)
		portalAppID := s.portalAppIDValue()
		if len(portalAppID) > 0 {
			var portalBytes []byte
			if err := obj.Call(`org.kde.KWallet.readEntry`, 0, handle, `xdg-desktop-portal`, portalAppID, appID).Store(&portalBytes); err == nil && len(portalBytes) > 0 {
				return portalBytes, nil
			}
		}
	}

	return nil, errors.New(`kwallet: password not found`)
}
