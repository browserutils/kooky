package lynx

import (
	"errors"
	"os"

	"github.com/zellyn/kooky"
)

type lynxCookieStore struct {
	filename         string
	file             *os.File
	browser          string
	isDefaultProfile bool
}

func (s *lynxCookieStore) FilePath() string {
	if s == nil {
		return ``
	}
	return s.filename
}
func (s *lynxCookieStore) Browser() string {
	if s == nil {
		return ``
	}
	return s.browser
}
func (s *lynxCookieStore) Profile() string {
	return ``
}
func (s *lynxCookieStore) IsDefaultProfile() bool {
	return true
}

var _ kooky.CookieStore = (*lynxCookieStore)(nil)

func (s *lynxCookieStore) open() error {
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

func (s *lynxCookieStore) Close() error {
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
