package uzbl

import (
	"errors"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/netscape"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s := &uzblCookieStore{filename: filename}
	defer s.Close()

	return s.ReadCookies(filters...)
}

func (s *uzblCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.open(); err != nil {
		return nil, err
	} else if s.file == nil {
		return nil, errors.New(`file is nil`)
	}

	cookies, _, err := netscape.ReadCookies(s.file, filters...)

	return cookies, err
}
