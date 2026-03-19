package chromium

import (
	"context"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/chrome"
	"github.com/browserutils/kooky/internal/cookies"
)

// KeyringConfig configures how the cookie store retrieves
// the decryption key from the system keyring.
//
// All fields are optional. When empty, defaults are derived
// from the browser name (e.g. "Chromium" → "Chromium Safe Storage").
//
// See the exported Keyring* variables for known browser configurations.
type KeyringConfig struct {
	Account     string // keychain account, e.g. "Chrome", "Microsoft Edge"
	Storage     string // keychain entry, e.g. "Chrome Safe Storage" (derived from Account if empty)
	Application string // secret service / kwallet app attr, e.g. "chrome" (derived from Account if empty)
	PortalAppID string // xdg-desktop-portal app ID, e.g. "org.chromium.Chromium"
}

// Known keyring configurations for Chromium-based browsers
// that do not have their own package.
var (
	// linux: $XDG_CONFIG_HOME/vivaldi/Default/Cookies
	KeyringConfigVivaldiLinux  = &KeyringConfig{Account: `Chrome`, PortalAppID: `com.vivaldi.Vivaldi`}
	KeyringConfigVivaldiDarwin = &KeyringConfig{Account: `Vivaldi`}

	// Application may be "arc" or "chrome" depending on version
	KeyringConfigArc = &KeyringConfig{Account: `Arc`}
)

func ReadCookies(ctx context.Context, filename string, keyring *KeyringConfig, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	return cookies.SingleRead(cookieStoreFunc(keyring), filename, filters...).ReadAllCookies(ctx)
}

func TraverseCookies(filename string, keyring *KeyringConfig, filters ...kooky.Filter) kooky.CookieSeq {
	return cookies.SingleRead(cookieStoreFunc(keyring), filename, filters...)
}

// CookieStore has to be closed with CookieStore.Close() after use.
func CookieStore(filename string, keyring *KeyringConfig, filters ...kooky.Filter) (kooky.CookieStore, error) {
	return cookieStore(filename, keyring, filters...)
}

func cookieStoreFunc(keyring *KeyringConfig) func(filename string, filters ...kooky.Filter) (*cookies.CookieJar, error) {
	return func(filename string, filters ...kooky.Filter) (*cookies.CookieJar, error) {
		return cookieStore(filename, keyring, filters...)
	}
}

func cookieStore(filename string, keyring *KeyringConfig, filters ...kooky.Filter) (*cookies.CookieJar, error) {
	s := &chrome.CookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `chromium`

	if keyring != nil {
		s.SetSafeStorage(keyring.Account, keyring.Storage, keyring.Application)
		if len(keyring.PortalAppID) > 0 {
			s.SetPortalAppID(keyring.PortalAppID)
		}
	}

	return cookies.NewCookieJar(s, filters...), nil
}
