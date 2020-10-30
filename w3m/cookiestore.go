package w3m

import (
	"errors"
	"os"

	"github.com/zellyn/kooky"
)

type w3mCookieStore struct {
	filename         string
	file             *os.File
	browser          string
	isDefaultProfile bool
}

func (s *w3mCookieStore) FilePath() string {
	if s == nil {
		return ``
	}
	return s.filename
}
func (s *w3mCookieStore) Browser() string {
	if s == nil {
		return ``
	}
	return s.browser
}
func (s *w3mCookieStore) Profile() string {
	return ``
}
func (s *w3mCookieStore) IsDefaultProfile() bool {
	return true
}

var _ kooky.CookieStore = (*w3mCookieStore)(nil)

func (s *w3mCookieStore) open() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.file != nil {
		s.file.Seek(0, 0)
		return nil
	}

	f, err := os.Open(s.filename)
	if err != nil {
		return err
	}
	s.file = f

	return nil
}

func (s *w3mCookieStore) Close() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.file == nil {
		return nil
	}
	err := s.file.Close()
	if err == nil {
		s.file = nil
	}

	return err
}
