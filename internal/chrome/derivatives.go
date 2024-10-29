package chrome

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type safeStorage struct {
	account string
	name    string
}

func (s *CookieStore) SetSafeStorage(account, name string) {
	if s == nil {
		return
	}
	if len(account) == 0 && len(s.BrowserStr) > 0 {
		account = cases.Title(language.English, cases.Compact).String(s.BrowserStr)
	}
	if len(name) == 0 {
		name = account + ` Safe Storage`
	}
	s.storage.account = account
	s.storage.name = name
}

func (s *CookieStore) safeStorageName() string {
	if s == nil {
		return ``
	}
	if len(s.storage.name) == 0 {
		s.SetSafeStorage(``, ``)
	}
	return s.storage.name
}

func (s *CookieStore) safeStorageAccount() string {
	if s == nil {
		return ``
	}
	if len(s.storage.name) == 0 {
		s.SetSafeStorage(``, ``)
	}
	return s.storage.account
}

/*
known "Safe Storage" string combinations:
account          |   name
----------------------------
Chrome           |   Chrome Safe Storage
Microsoft Edge   |   Microsoft Edge Safe Storage
*/
