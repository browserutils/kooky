package cookies

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/iterx"
	"github.com/browserutils/kooky/internal/utils"
)

// kooky.CookieStore without http.CookieJar and SubJar()
type CookieStore interface {
	TraverseCookies(...kooky.Filter) kooky.CookieSeq
	Browser() string
	Profile() string
	IsDefaultProfile() bool
	FilePath() string
	Close() error
}

/*
DefaultCookieStore implements most of the kooky.CookieStore interface except for the TraverseCookies method
func (s *DefaultCookieStore) TraverseCookies(filters ...kooky.Filter) kooky.CookieSeq

DefaultCookieStore also provides an Open() method
*/
type DefaultCookieStore struct {
	FileNameStr          string
	File                 *os.File
	BrowserStr           string
	ProfileStr           string
	OSStr                string
	IsDefaultProfileBool bool
}

func (s *DefaultCookieStore) FilePath() string {
	if s == nil {
		return ``
	}
	return s.FileNameStr
}
func (s *DefaultCookieStore) Browser() string {
	if s == nil {
		return ``
	}
	return s.BrowserStr
}
func (s *DefaultCookieStore) Profile() string {
	if s == nil {
		return ``
	}
	return s.ProfileStr
}
func (s *DefaultCookieStore) IsDefaultProfile() bool {
	return s != nil && s.IsDefaultProfileBool
}

func (s *DefaultCookieStore) Open() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.File != nil {
		s.File.Seek(0, io.SeekStart)
		return nil
	}
	if len(s.FileNameStr) < 1 {
		return nil
	}

	f, err := utils.OpenFile(s.FileNameStr)
	if err != nil {
		return err
	}
	s.File = f

	return nil
}

func (s *DefaultCookieStore) Close() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.File == nil {
		return nil
	}
	err := s.File.Close()
	if err == nil {
		s.File = nil
	}

	return err
}

type JarCreator func(filename string, filters ...kooky.Filter) (*CookieJar, error)

func SingleRead(jarCr JarCreator, filename string, filters ...kooky.Filter) kooky.CookieSeq {
	st, err := jarCr(filename, filters...)
	if err != nil {
		return iterx.ErrCookieSeq(err)
	}
	return ReadCookiesClose(st, filters...)
}

func ReadCookiesClose(store CookieStore, filters ...kooky.Filter) kooky.CookieSeq {
	if store == nil {
		return iterx.ErrCookieSeq(errors.New(`nil cookie store`))
	}
	seq := func(yield func(*kooky.Cookie, error) bool) {
		defer func() {
			if err := store.Close(); err != nil {
				yield(nil, err)
			}
		}()
		for cookie, err := range store.TraverseCookies(filters...) {
			if !iterx.CookieFilterYield(context.Background(), cookie, err, yield, filters...) {
				return
			}
		}

	}
	return seq
}
