package dillo

import (
	"errors"
	"os"

	"github.com/zellyn/kooky"
)

type dilloCookieStore struct {
	filename         string
	file             *os.File
	browser          string
	isDefaultProfile bool
}

func (s *dilloCookieStore) FilePath() string {
	if s == nil {
		return ``
	}
	return s.filename
}
func (s *dilloCookieStore) Browser() string {
	if s == nil {
		return ``
	}
	return s.browser
}
func (s *dilloCookieStore) Profile() string {
	return ``
}
func (s *dilloCookieStore) IsDefaultProfile() bool {
	return true
}

var _ kooky.CookieStore = (*dilloCookieStore)(nil)

func (s *dilloCookieStore) open() error {
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

func (s *dilloCookieStore) Close() error {
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
