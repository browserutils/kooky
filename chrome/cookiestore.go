package chrome

import (
	"errors"

	"github.com/zellyn/kooky"

	"github.com/go-sqlite/sqlite3"
)

type chromeCookieStore struct {
	filename         string
	database         *sqlite3.DbFile
	keyringPassword  []byte
	password         []byte
	os               string
	browser          string
	profile          string
	isDefaultProfile bool
	decryptionMethod func(data, password []byte) ([]byte, error)
}

func (s *chromeCookieStore) FilePath() string {
	if s == nil {
		return ``
	}
	return s.filename
}
func (s *chromeCookieStore) Browser() string {
	if s == nil {
		return ``
	}
	return s.browser
}
func (s *chromeCookieStore) Profile() string {
	if s == nil {
		return ``
	}
	return s.profile
}
func (s *chromeCookieStore) IsDefaultProfile() bool {
	if s == nil {
		return false
	}
	return s.isDefaultProfile
}

var _ kooky.CookieStore = (*chromeCookieStore)(nil)

func (s *chromeCookieStore) open() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.database != nil {
		return nil
	}

	db, err := sqlite3.Open(s.filename)
	if err != nil {
		return err
	}
	s.database = db

	return nil
}

func (s *chromeCookieStore) Close() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.database == nil {
		return nil
	}
	err := s.database.Close()
	if err == nil {
		s.database = nil
	}

	return err
}
