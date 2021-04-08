package opera

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/chrome"

	"github.com/go-sqlite/sqlite3"
)

type operaCookieStore struct {
	chrome.CookieStore
}

var _ kooky.CookieStore = (*operaCookieStore)(nil)

func (s *operaCookieStore) Open() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.File != nil {
		s.File.Seek(0, 0)
		return nil
	}
	if len(s.FileNameStr) < 1 {
		return nil
	}

	// TODO use file type detection

	if filepath.Base(s.FileNameStr) == `cookies4.dat` {
		f, err := os.Open(s.FileNameStr)
		if err != nil {
			return err
		}
		s.File = f
	} else {
		db, err := sqlite3.Open(s.FileNameStr)
		if err != nil {
			return err
		}
		s.Database = db
	}

	return nil
}

func (s *operaCookieStore) Close() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}

	var err, errFile, errDB error

	if s.File != nil {
		errFile = s.File.Close()
		s.File = nil
	}
	if s.Database != nil {
		errDB = s.Database.Close()
		s.File = nil
	}

	if errFile != nil && errDB == nil {
		err = errFile
	} else if errFile == nil && errDB != nil {
		err = errDB
	} else if errFile != nil && errDB != nil {
		err = fmt.Errorf("os.File.Close() error \"%v\" and github.com/go-sqlite/sqlite3.DbFile.Close() error \"%v\" occurred", errFile, errDB)
	}

	return err
}
