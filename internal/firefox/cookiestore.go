package firefox

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/utils"
	"github.com/go-sqlite/sqlite3"
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

	f, err := utils.OpenFile(s.FileNameStr)
	if err != nil {
		return err
	}
	db, err := sqlite3.OpenFrom(f)
	if err != nil {
		return err
	}
	s.Database = db

	contFileName := filepath.Join(filepath.Dir(s.FileNameStr), `containers.json`)
	s.contFile, _ = utils.OpenFile(contFileName)

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
