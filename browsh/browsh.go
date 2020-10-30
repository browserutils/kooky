// Browsh Browser
package browsh

import (
	"errors"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/firefox"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s := &browshCookieStore{
		filename: filename,
		browser:  `browsh`,
	}
	defer s.Close()

	return s.ReadCookies(filters...)
}

func (s *browshCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
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
		IsDefaultProfile: s.isDefaultProfile,
	}

	cookies, err := cookieStore.ReadCookies(filters...)

	s.filename = cookieStore.Filename
	s.database = cookieStore.Database
	s.browser = cookieStore.Browser
	s.isDefaultProfile = cookieStore.IsDefaultProfile

	return cookies, err
}
