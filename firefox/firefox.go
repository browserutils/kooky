package firefox

import (
	"errors"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/firefox"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s := &firefoxCookieStore{
		filename: filename,
		browser:  `firefox`,
	}
	defer s.Close()

	return s.ReadCookies(filters...)
}

func (s *firefoxCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.open(); err != nil {
		return nil, err
	} else if s.database == nil {
		return nil, errors.New(`database is nil`)
	}

	cookieStore := &firefox.CookieStore{
		Filename:         s.filename,
		Database:         s.database,
		Browser:          s.browser,
		Profile:          s.profile,
		IsDefaultProfile: s.isDefaultProfile,
	}

	cookies, err := cookieStore.ReadCookies(filters...)

	s.filename = cookieStore.Filename
	s.database = cookieStore.Database
	s.browser = cookieStore.Browser
	s.profile = cookieStore.Profile
	s.isDefaultProfile = cookieStore.IsDefaultProfile

	return cookies, err
}
