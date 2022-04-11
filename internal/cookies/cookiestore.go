package cookies

import (
	"errors"
	"io"
	"os"

	"github.com/zellyn/kooky"
)

// kooky.CookieStore without http.CookieJar and SubJar()
type CookieStore interface {
	ReadCookies(...kooky.Filter) ([]*kooky.Cookie, error)
	Browser() string
	Profile() string
	IsDefaultProfile() bool
	FilePath() string
	Close() error
}

/*
DefaultCookieStore implements most of the kooky.CookieStore interface except for the ReadCookies method
func (s *DefaultCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error)

DefaultCookieStore also provides an Open() method
*/
type DefaultCookieStore struct {
	FileNameStr          string
	File                 *os.File
	BrowserStr           string
	ProfileStr           string
	OSStr                string
	IsDefaultProfileBool bool
}

func (s *DefaultCookieStore) FilePath() string {
	if s == nil {
		return ``
	}
	return s.FileNameStr
}
func (s *DefaultCookieStore) Browser() string {
	if s == nil {
		return ``
	}
	return s.BrowserStr
}
func (s *DefaultCookieStore) Profile() string {
	if s == nil {
		return ``
	}
	return s.ProfileStr
}
func (s *DefaultCookieStore) IsDefaultProfile() bool {
	return s != nil && s.IsDefaultProfileBool
}

func (s *DefaultCookieStore) Open() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.File != nil {
		s.File.Seek(0, io.SeekStart)
		return nil
	}
	if len(s.FileNameStr) < 1 {
		return nil
	}

	f, err := os.Open(s.FileNameStr)
	if err != nil {
		return err
	}
	s.File = f

	return nil
}

func (s *DefaultCookieStore) Close() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.File == nil {
		return nil
	}
	err := s.File.Close()
	if err == nil {
		s.File = nil
	}

	return err
}
