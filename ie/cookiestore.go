package ie

import (
	"errors"
	"os"

	"github.com/zellyn/kooky"
)

type ieCookieStore struct {
	filename         string
	file             *os.File
	browser          string
	isDefaultProfile bool
}

func (s *ieCookieStore) FilePath() string {
	if s == nil {
		return ``
	}
	return s.filename
}
func (s *ieCookieStore) Browser() string {
	if s == nil {
		return ``
	}
	return s.browser
}
func (s *ieCookieStore) Profile() string {
	return ``
}
func (s *ieCookieStore) IsDefaultProfile() bool {
	return true
}

var _ kooky.CookieStore = (*ieCookieStore)(nil)

func (s *ieCookieStore) open() error {
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

func (s *ieCookieStore) Close() error {
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
