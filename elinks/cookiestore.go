package elinks

import (
	"errors"
	"os"

	"github.com/zellyn/kooky"
)

type elinksCookieStore struct {
	filename         string
	file             *os.File
	browser          string
	isDefaultProfile bool
}

func (s *elinksCookieStore) FilePath() string {
	if s == nil {
		return ``
	}
	return s.filename
}
func (s *elinksCookieStore) Browser() string {
	if s == nil {
		return ``
	}
	return s.browser
}
func (s *elinksCookieStore) Profile() string {
	return ``
}
func (s *elinksCookieStore) IsDefaultProfile() bool {
	return true
}

var _ kooky.CookieStore = (*elinksCookieStore)(nil)

func (s *elinksCookieStore) open() error {
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

func (s *elinksCookieStore) Close() error {
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
