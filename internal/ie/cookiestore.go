package ie

import (
	"errors"
	"io"
	"os"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/cookies"
	"github.com/zellyn/kooky/internal/utils"

	"www.velocidex.com/golang/go-ese/parser"
)

type CookieStore struct {
	cookies.CookieStore
}

var _ cookies.CookieStore = (*CookieStore)(nil)

func (s *CookieStore) Open() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if o, ok := s.CookieStore.(interface{ Open() error }); ok {
		return o.Open()
	}
	return nil
}

func (s *CookieStore) Close() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	return s.CookieStore.Close()
}

// TODO might need temporary copy of the file
type IECacheCookieStore struct {
	cookies.DefaultCookieStore
}

var _ cookies.CookieStore = (*IECacheCookieStore)(nil)

type ESECookieStore struct {
	cookies.DefaultCookieStore
	ESECatalog *parser.Catalog
}

var _ cookies.CookieStore = (*ESECookieStore)(nil)

func (s *ESECookieStore) Open() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.File != nil {
		s.File.Seek(0, io.SeekStart)
		return nil
	}
	if s.ESECatalog != nil {
		return nil
	}
	// TODO: create temporary copy of the file on Windows - a service on Windows has a permanent lock on it
	// TODO: remove temporary copy in Close()
	if f, err := os.Open(s.FileNameStr); err != nil {
		return err
	} else {
		s.File = f
	}
	if _, err := s.File.Seek(0, io.SeekStart); err != nil {
		return err
	}

	ese_ctx, err := parser.NewESEContext(s.File)
	if err != nil {
		return err
	}

	catalog, err := parser.ReadCatalog(ese_ctx)
	if err != nil {
		return err
	}

	s.ESECatalog = catalog

	return nil
}

func (s *ESECookieStore) Close() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}

	var err error
	if s.File != nil {
		err = s.File.Close()
		s.File = nil
	}
	s.ESECatalog = nil

	return err
}

func GetCookieStore(filename, browser string, m map[string]func(f *os.File, s *CookieStore, browser string), filters ...kooky.Filter) (*cookies.CookieJar, error) {
	var s CookieStore

	f, typ, err := utils.DetectFileType(filename)
	if err != nil {
		return nil, err
	}
	if m != nil {
		for name, fn := range m {
			if f != nil && typ == name {
				fn(f, &s, browser)
				goto end
			}
		}
	}
	switch typ {
	case `ie_cache`:
		t := &IECacheCookieStore{}
		t.File = f
		t.FileNameStr = filename
		t.BrowserStr = browser
		s.CookieStore = t
	case `ese`:
		e := &ESECookieStore{}
		e.File = f
		e.FileNameStr = filename
		e.BrowserStr = browser
		s.CookieStore = e
	default:
		f.Close()
		return nil, errors.New(`unknown file type`)
	}
end:

	return &cookies.CookieJar{CookieStore: &s}, nil
}
