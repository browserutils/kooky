package firefox

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/go-sqlite/sqlite3"
	"github.com/zellyn/kooky/internal/cookies"
)

type CookieStore struct {
	cookies.DefaultCookieStore
	Database   *sqlite3.DbFile
	Containers map[int]string
	contFile   *os.File
}

var _ cookies.CookieStore = (*CookieStore)(nil)

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

	contFileName := filepath.Join(filepath.Dir(s.FileNameStr), `containers.json`)
	s.contFile, _ = os.Open(contFileName)

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
	if s.contFile != nil {
		errCont := s.contFile.Close()
		if errCont != nil && err == nil {
			err = errCont
		}
	}

	return err
}
