package ladybird

import (
	"errors"
	"os"

	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/utils"
	"github.com/go-sqlite/sqlite3"
)

type ladybirdCookieStore struct {
	cookies.DefaultCookieStore
	Database *sqlite3.DbFile
	dbFile   *os.File
}

var _ cookies.CookieStore = (*ladybirdCookieStore)(nil)

func (s *ladybirdCookieStore) Open() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.Database != nil {
		return nil
	}

	f, err := utils.OpenFile(s.FileNameStr)
	if err != nil {
		return err
	}
	db, err := sqlite3.OpenFrom(f)
	if err != nil {
		f.Close()
		return err
	}
	s.Database = db
	s.dbFile = f

	return nil
}

func (s *ladybirdCookieStore) Close() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.Database == nil {
		return nil
	}
	err := s.Database.Close()
	if s.dbFile != nil {
		if errDB := s.dbFile.Close(); errDB != nil && err == nil {
			err = errDB
		}
	}
	if err == nil {
		s.Database = nil
	}

	return err
}
