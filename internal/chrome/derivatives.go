package chrome

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type safeStorage struct {
	account     string // e.g. "Chromium", "Chrome", "Microsoft Edge"
	name        string // e.g. "Chromium Safe Storage"
	application string // Secret Service application attr, e.g. "chromium", "chrome"
	portalAppID string // xdg-desktop-portal app ID, e.g. "com.vivaldi.Vivaldi"
}

func (s *CookieStore) SetSafeStorage(account, name, application string) {
	if s == nil {
		return
	}
	if len(account) == 0 && len(s.BrowserStr) > 0 {
		account = cases.Title(language.English, cases.Compact).String(s.BrowserStr)
	}
	if len(name) == 0 {
		name = account + ` Safe Storage`
	}
	if len(application) == 0 {
		application = strings.ToLower(account)
	}
	s.storage.account = account
	s.storage.name = name
	s.storage.application = application
}

func (s *CookieStore) safeStorageName() string {
	if s == nil {
		return ``
	}
	if len(s.storage.name) == 0 {
		s.SetSafeStorage(``, ``, ``)
	}
	return s.storage.name
}

func (s *CookieStore) safeStorageAccount() string {
	if s == nil {
		return ``
	}
	if len(s.storage.name) == 0 {
		s.SetSafeStorage(``, ``, ``)
	}
	return s.storage.account
}

func (s *CookieStore) safeStorageApplication() string {
	if s == nil {
		return ``
	}
	if len(s.storage.name) == 0 {
		s.SetSafeStorage(``, ``, ``)
	}
	return s.storage.application
}

func (s *CookieStore) SetPortalAppID(id string) {
	if s == nil {
		return
	}
	s.storage.portalAppID = id
}

func (s *CookieStore) portalAppIDValue() string {
	if s == nil {
		return ``
	}
	return s.storage.portalAppID
}
