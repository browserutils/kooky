package edge

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zellyn/kooky"

	"github.com/go-sqlite/sqlite3"
)

type edgeCookieStore struct {
	filename         string
	file             *os.File
	database         *sqlite3.DbFile
	keyringPassword  []byte
	password         []byte
	browser          string
	profile          string
	os               string
	isDefaultProfile bool
	decryptionMethod func(data, password []byte) ([]byte, error)
	// os            string
}

func (s *edgeCookieStore) FilePath() string {
	if s == nil {
		return ``
	}
	return s.filename
}
func (s *edgeCookieStore) Browser() string {
	if s == nil {
		return ``
	}
	return s.browser
}
func (s *edgeCookieStore) Profile() string {
	if s == nil {
		return ``
	}
	return s.profile
}
func (s *edgeCookieStore) IsDefaultProfile() bool {
	if s == nil {
		return false
	}
	return s.isDefaultProfile
}

var _ kooky.CookieStore = (*edgeCookieStore)(nil)

func (s *edgeCookieStore) open() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.file != nil {
		s.file.Seek(0, 0)
		return nil
	}

	// TODO use file type detection

	if filepath.Base(s.filename) == `cookies4.dat` {
		f, err := os.Open(s.filename)
		if err != nil {
			return err
		}
		s.file = f
	} else {
		db, err := sqlite3.Open(s.filename)
		if err != nil {
			return err
		}
		s.database = db
	}

	return nil
}

func (s *edgeCookieStore) Close() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}

	var err, errFile, errDB error

	if s.file != nil {
		errFile = s.file.Close()
		s.file = nil
	}
	if s.database != nil {
		errDB = s.database.Close()
		s.file = nil
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
