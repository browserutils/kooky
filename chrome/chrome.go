package chrome

import (
	"errors"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/chrome"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s := &chromeCookieStore{
		filename: filename,
		browser:  `chrome|chromium`, // TODO
	}
	defer s.Close()

	return s.ReadCookies(filters...)
}

func (s *chromeCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.open(); err != nil {
		return nil, err
	} else if s.database == nil {
		return nil, errors.New(`database is nil`)
	}

	cookieStore := &chrome.CookieStore{
		Filename:         s.filename,
		Database:         s.database,
		KeyringPassword:  s.keyringPassword,
		Password:         s.password,
		OS:               s.os,
		Browser:          s.browser,
		Profile:          s.profile,
		IsDefaultProfile: s.isDefaultProfile,
		DecryptionMethod: s.decryptionMethod,
	}

	cookies, err := cookieStore.ReadCookies(filters...)

	s.filename = cookieStore.Filename
	s.database = cookieStore.Database
	s.keyringPassword = cookieStore.KeyringPassword
	s.password = cookieStore.Password
	s.os = cookieStore.OS
	s.browser = cookieStore.Browser
	s.profile = cookieStore.Profile
	s.isDefaultProfile = cookieStore.IsDefaultProfile
	s.decryptionMethod = cookieStore.DecryptionMethod

	return cookies, err
}

// returns the previous password for later restoration
// used in tests
func (s *chromeCookieStore) setKeyringPassword(password []byte) []byte {
	if s == nil {
		return nil
	}
	oldPassword := s.keyringPassword
	s.keyringPassword = password
	return oldPassword
}
