package uzbl

import (
	"errors"
	"os"

	"github.com/zellyn/kooky"
)

type uzblCookieStore struct {
	filename         string
	file             *os.File
	browser          string
	isDefaultProfile bool
}

func (s *uzblCookieStore) FilePath() string {
	if s == nil {
		return ``
	}
	return s.filename
}
func (s *uzblCookieStore) Browser() string {
	if s == nil {
		return ``
	}
	return s.browser
}
func (s *uzblCookieStore) Profile() string {
	return ``
}
func (s *uzblCookieStore) IsDefaultProfile() bool {
	return true
}

var _ kooky.CookieStore = (*uzblCookieStore)(nil)

func (s *uzblCookieStore) open() error {
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

func (s *uzblCookieStore) Close() error {
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
