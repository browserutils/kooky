package cookies

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"

	"golang.org/x/net/publicsuffix"

	"github.com/zellyn/kooky"
)

var (
	_ http.CookieJar    = (*CookieJar)(nil)
	_ kooky.CookieStore = (*CookieJar)(nil)
)

type CookieJar struct {
	*cookiejar.Jar
	init    sync.Once
	initErr error
	CookieStore
}

func (s *CookieJar) Cookies(u *url.URL) []*http.Cookie {
	if s == nil || s.CookieStore == nil || s.initErr != nil {
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

func (s *CookieJar) InitJar() error {
	if s == nil {
		return errors.New(`nil receiver`)
	}
	if s.CookieStore == nil {
		return errors.New(`no cookie store set`)
	}
	s.init.Do(func() {
		kookies, err := s.CookieStore.ReadCookies()
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

func (s *CookieJar) SubJar(filters ...kooky.Filter) (http.CookieJar, error) {
	if s == nil {
		return nil, errors.New(`nil receiver`)
	}
	if s.CookieStore == nil {
		return nil, errors.New(`no cookie store set`)
	}
	kookies, err := s.CookieStore.ReadCookies(filters...)
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
