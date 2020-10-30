package browsh

import (
	"errors"

	"github.com/zellyn/kooky"

	"github.com/go-sqlite/sqlite3"
)

type browshCookieStore struct {
	filename         string
	database         *sqlite3.DbFile
	browser          string
	isDefaultProfile bool
}

func (s *browshCookieStore) FilePath() string {
	if s == nil {
		return ``
	}
	return s.filename
}
func (s *browshCookieStore) Browser() string {
	if s == nil {
		return ``
	}
	return s.browser
}
func (s *browshCookieStore) Profile() string {
	return ``
}
func (s *browshCookieStore) IsDefaultProfile() bool {
	if s == nil {
		return false
	}
	return s.isDefaultProfile
}

var _ kooky.CookieStore = (*browshCookieStore)(nil)

func (s *browshCookieStore) open() error {
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

func (s *browshCookieStore) Close() error {
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
