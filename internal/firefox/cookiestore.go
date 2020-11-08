package firefox

import (
	"errors"

	"github.com/go-sqlite/sqlite3"
	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal"
)

type CookieStore struct {
	internal.DefaultCookieStore
	Database *sqlite3.DbFile
}

func (s *CookieStore) Open() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.Database != nil {
		return nil
	}

	db, err := sqlite3.Open(s.FileNameStr)
	if err != nil {
		return err
	}
	s.Database = db

	return nil
}

func (s *CookieStore) Close() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.Database == nil {
		return nil
	}
	err := s.Database.Close()
	if err == nil {
		s.Database = nil
	}

	return err
}

var _ kooky.CookieStore = (*CookieStore)(nil)
