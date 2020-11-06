package safari

import (
	"errors"
	"os"

	"github.com/zellyn/kooky"
)

type safariCookieStore struct {
	filename         string
	file             *os.File
	keyringPassword  []byte
	password         []byte
	browser          string
	profile          string
	isDefaultProfile bool
}

func (s *safariCookieStore) FilePath() string {
	if s == nil {
		return ``
	}
	return s.filename
}
func (s *safariCookieStore) Browser() string {
	if s == nil {
		return ``
	}
	return s.browser
}
func (s *safariCookieStore) Profile() string {
	if s == nil {
		return ``
	}
	return s.profile
}
func (s *safariCookieStore) IsDefaultProfile() bool {
	if s == nil {
		return false
	}
	return s.isDefaultProfile
}

var _ kooky.CookieStore = (*safariCookieStore)(nil)

func (s *safariCookieStore) open() error {
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

func (s *safariCookieStore) Close() error {
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
