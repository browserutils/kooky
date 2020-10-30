package firefox

import (
	"errors"

	"github.com/zellyn/kooky"

	"github.com/go-sqlite/sqlite3"
)

type firefoxCookieStore struct {
	filename         string
	database         *sqlite3.DbFile
	browser          string
	profile          string
	isDefaultProfile bool
}

func (s *firefoxCookieStore) FilePath() string {
	if s == nil {
		return ``
	}
	return s.filename
}
func (s *firefoxCookieStore) Browser() string {
	if s == nil {
		return ``
	}
	return s.browser
}
func (s *firefoxCookieStore) Profile() string {
	if s == nil {
		return ``
	}
	return s.profile
}
func (s *firefoxCookieStore) IsDefaultProfile() bool {
	if s == nil {
		return false
	}
	return s.isDefaultProfile
}

var _ kooky.CookieStore = (*firefoxCookieStore)(nil)

func (s *firefoxCookieStore) open() error {
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

func (s *firefoxCookieStore) Close() error {
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
