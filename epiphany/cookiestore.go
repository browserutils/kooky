package epiphany

import (
	"errors"

	"github.com/zellyn/kooky"

	"github.com/go-sqlite/sqlite3"
)

type epiphanyCookieStore struct {
	filename         string
	database         *sqlite3.DbFile
	browser          string
	isDefaultProfile bool
}

func (s *epiphanyCookieStore) FilePath() string {
	if s == nil {
		return ``
	}
	return s.filename
}
func (s *epiphanyCookieStore) Browser() string {
	if s == nil {
		return ``
	}
	return s.browser
}
func (s *epiphanyCookieStore) Profile() string {
	return ``
}
func (s *epiphanyCookieStore) IsDefaultProfile() bool {
	if s == nil {
		return false
	}
	return s.isDefaultProfile
}

var _ kooky.CookieStore = (*epiphanyCookieStore)(nil)

func (s *epiphanyCookieStore) open() error {
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

func (s *epiphanyCookieStore) Close() error {
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
