package netscape

import (
	"errors"
	"os"

	"github.com/zellyn/kooky"
)

type netscapeCookieStore struct {
	filename         string
	file             *os.File
	browser          string
	profile          string
	isDefaultProfile bool
	isStrict         bool
}

func (s *netscapeCookieStore) FilePath() string {
	if s == nil {
		return ``
	}
	return s.filename
}
func (s *netscapeCookieStore) Browser() string {
	if s == nil {
		return ``
	}
	return s.browser
}
func (s *netscapeCookieStore) Profile() string {
	if s == nil {
		return ``
	}
	return s.profile
}
func (s *netscapeCookieStore) IsDefaultProfile() bool {
	if s == nil {
		return false
	}
	return s.isDefaultProfile
}

var _ kooky.CookieStore = (*netscapeCookieStore)(nil)

func (s *netscapeCookieStore) open() error {
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

func (s *netscapeCookieStore) Close() error {
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
