package opera

import (
	"errors"
	"io"
	"os"

	"github.com/zellyn/kooky/internal/cookies"
)

type operaCookieStore struct {
	cookies.CookieStore
}

var _ cookies.CookieStore = (*operaCookieStore)(nil)

type operaPrestoCookieStore struct {
	cookies.DefaultCookieStore
}

var _ cookies.CookieStore = (*operaPrestoCookieStore)(nil)

func (s *operaPrestoCookieStore) Open() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.File != nil {
		s.File.Seek(0, io.SeekStart)
		return nil
	}
	if len(s.FileNameStr) == 0 {
		return nil
	}

	f, err := os.Open(s.FileNameStr)
	if err != nil {
		return err
	}
	s.File = f

	return nil
}

func (s *operaPrestoCookieStore) Close() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}

	var err error

	if s.File != nil {
		err = s.File.Close()
		s.File = nil
	}

	return err
}
