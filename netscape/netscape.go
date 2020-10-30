package netscape

import (
	"errors"
	"os"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/netscape"
)

func ReadCookies(filename string, filters ...kooky.Filter) (c []*kooky.Cookie, strict bool, e error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, false, err
	}
	defer f.Close()

	return netscape.ReadCookies(f, filters...)
}

func (s *netscapeCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.open(); err != nil {
		return nil, err
	} else if s.file == nil {
		return nil, errors.New(`file is nil`)
	}

	cookies, isStrict, err := netscape.ReadCookies(s.file, filters...)
	s.isStrict = isStrict

	return cookies, err
}
