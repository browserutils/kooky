package chrome

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type safeStorage struct {
	account     string // e.g. "Chromium", "Chrome", "Microsoft Edge"
	storage     string // e.g. "Chromium Safe Storage"
	application string // Secret Service application attr, e.g. "chromium", "chrome"
	portalAppID string // xdg-desktop-portal app ID, e.g. "com.vivaldi.Vivaldi"
}

func (s *CookieStore) SetSafeStorage(account, storage, application string) {
	if s == nil {
		return
	}
	if len(account) == 0 && len(s.BrowserStr) > 0 {
		account = cases.Title(language.English, cases.Compact).String(s.BrowserStr)
	}
	if len(storage) == 0 {
		storage = account + ` Safe Storage`
	}
	if len(application) == 0 {
		application = strings.ToLower(account)
	}
	s.keyringConfig.account = account
	s.keyringConfig.storage = storage
	s.keyringConfig.application = application
}

func (s *CookieStore) SetPortalAppID(id string) {
	if s == nil {
		return
	}
	s.keyringConfig.portalAppID = id
}

func (s *CookieStore) ensureKeyring() {
	if s != nil && len(s.keyringConfig.storage) == 0 {
		s.SetSafeStorage(``, ``, ``)
	}
}

func (s *CookieStore) safeStorageAccount() string {
	if s == nil {
		return ``
	}
	s.ensureKeyring()
	return s.keyringConfig.account
}

func (s *CookieStore) safeStorageName() string {
	if s == nil {
		return ``
	}
	s.ensureKeyring()
	return s.keyringConfig.storage
}

func (s *CookieStore) safeStorageApplication() string {
	if s == nil {
		return ``
	}
	s.ensureKeyring()
	return s.keyringConfig.application
}

func (s *CookieStore) portalAppIDValue() string {
	if s == nil {
		return ``
	}
	return s.keyringConfig.portalAppID
}
