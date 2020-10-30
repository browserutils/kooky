package konqueror

import (
	"errors"
	"os"

	"github.com/zellyn/kooky"
)

type konquerorCookieStore struct {
	filename         string
	file             *os.File
	browser          string
	isDefaultProfile bool
}

func (s *konquerorCookieStore) FilePath() string {
	if s == nil {
		return ``
	}
	return s.filename
}
func (s *konquerorCookieStore) Browser() string {
	if s == nil {
		return ``
	}
	return s.browser
}
func (s *konquerorCookieStore) Profile() string {
	return ``
}
func (s *konquerorCookieStore) IsDefaultProfile() bool {
	return true
}

var _ kooky.CookieStore = (*konquerorCookieStore)(nil)

func (s *konquerorCookieStore) open() error {
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

func (s *konquerorCookieStore) Close() error {
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
