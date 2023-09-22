package epiphany

import (
	"errors"

	"github.com/browserutils/kooky/internal/cookies"
	"github.com/go-sqlite/sqlite3"
)

type epiphanyCookieStore struct {
	cookies.DefaultCookieStore
	Database *sqlite3.DbFile
}

var _ cookies.CookieStore = (*epiphanyCookieStore)(nil)

func (s *epiphanyCookieStore) Open() error {
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

func (s *epiphanyCookieStore) Close() error {
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
