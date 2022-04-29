package chrome

import (
	"errors"

	"github.com/go-sqlite/sqlite3"

	"github.com/zellyn/kooky/internal/cookies"
)

type CookieStore struct {
	cookies.DefaultCookieStore
	Database             *sqlite3.DbFile
	KeyringPasswordBytes []byte
	PasswordBytes        []byte
	DecryptionMethod     func(data, password []byte) ([]byte, error)
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

var _ cookies.CookieStore = (*CookieStore)(nil)

// returns the previous password for later restoration
// used in tests
func (s *CookieStore) SetKeyringPassword(password []byte) []byte {
	if s == nil {
		return nil
	}
	oldPassword := s.KeyringPasswordBytes
	s.KeyringPasswordBytes = password
	return oldPassword
}
