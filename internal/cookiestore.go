package internal

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"sync"

	"github.com/zellyn/kooky"
	"golang.org/x/net/publicsuffix"
)

var _ http.CookieJar = (*DefaultCookieStore)(nil)

type DefaultCookieStore struct {
	*cookiejar.Jar
	init    sync.Once
	initErr error
	store   kooky.CookieStore

	FileNameStr          string
	File                 *os.File
	BrowserStr           string
	ProfileStr           string
	OSStr                string
	IsDefaultProfileBool bool
}

/*
DefaultCookieStore implements most of the kooky.CookieStore interface except for the ReadCookies method
func (s *DefaultCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error)

DefaultCookieStore also provides an Open() method
*/

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
		s.File.Seek(0, 0)
		return nil
	}
	if len(s.FileNameStr) < 1 {
		return nil
	}

	f, err := os.Open(s.FileNameStr)
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

func (s *DefaultCookieStore) Cookies(u *url.URL) []*http.Cookie {
	if s == nil || s.store == nil {
		return nil
	}
	if err := s.InitJar(); err != nil {
		return nil
	}
	if s.Jar == nil {
		return nil
	}
	return s.Jar.Cookies(u)
}

func (s *DefaultCookieStore) InitJar() error {
	if s == nil {
		return errors.New(`nil receiver`)
	}
	if s.store == nil {
		return errors.New(`no cookie store set`)
	}
	s.init.Do(func() {
		kookies, err := s.store.ReadCookies()
		defer s.Close()
		if err != nil {
			s.initErr = err
			return
		}
		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		if err != nil {
			s.initErr = err
			return
		}
		s.Jar = jar
		cookies := kookies2cookies(kookies)
		setAllCookies(s, cookies)
	})

	return s.initErr
}

func (s *DefaultCookieStore) SubJar(filters ...kooky.Filter) (http.CookieJar, error) {
	if s == nil {
		return nil, errors.New(`nil receiver`)
	}
	if s.store == nil {
		return nil, errors.New(`no cookie store set`)
	}
	kookies, err := s.store.ReadCookies(filters...)
	defer s.Close()
	if err != nil {
		return nil, err
	}
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}
	cookies := kookies2cookies(kookies)
	setAllCookies(jar, cookies)

	return jar, nil
}

func SetCookieStore(d *DefaultCookieStore, s kooky.CookieStore) error {
	if d == nil {
		return errors.New(`received nil argument`)
	}
	d.store = s
	return nil
}

func kookies2cookies(kookies []*kooky.Cookie, filters ...kooky.Filter) []*http.Cookie {
	filteredKookies := kooky.FilterCookies(kookies, filters...)
	cookies := make([]*http.Cookie, len(filteredKookies))
	for _, k := range filteredKookies {
		cookies = append(cookies, &k.Cookie)
	}
	return cookies
}

func setAllCookies(jar http.CookieJar, cookies []*http.Cookie) {
	cookieMap := make(map[*url.URL][]*http.Cookie)

	for _, c := range cookies {
		if c == nil {
			continue
		}

		var scheme string
		if c.Secure {
			scheme = `https`
		} else {
			scheme = `http`
		}
		u := &url.URL{
			Scheme: scheme,
			Host:   c.Domain,
			Path:   c.Path,
		}

		if _, ok := cookieMap[u]; !ok {
			cookieMap[u] = []*http.Cookie{c}
			continue
		}
		cookieMap[u] = append(cookieMap[u], c)
	}

	for u, cs := range cookieMap {
		jar.SetCookies(u, cs)
	}
}
