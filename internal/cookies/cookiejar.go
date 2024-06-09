package cookies

import (
	"context"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"

	"golang.org/x/net/publicsuffix"

	"github.com/browserutils/kooky"
)

var (
	_ http.CookieJar    = (*CookieJar)(nil)
	_ kooky.CookieStore = (*CookieJar)(nil)
)

type CookieJar struct {
	*cookiejar.Jar
	init    sync.Once
	initErr error
	filters []kooky.Filter
	cookies kooky.Cookies // duplicate storage required for SubJar()
	CookieStore
}

func NewCookieJar(st CookieStore, filters ...kooky.Filter) *CookieJar {
	// TODO set struct fields of the CookieStore:
	// FileNameStr, BrowserStr, string; CookieStore CookieStore (inner), File *os.File
	return &CookieJar{
		filters:     filters,
		CookieStore: st,
	}
}

func (s *CookieJar) Cookies(u *url.URL) []*http.Cookie {
	var ret []*http.Cookie
	if s == nil || s.CookieStore == nil || s.initErr != nil {
		return ret
	}
	if err := s.InitJar(); err != nil {
		return ret
	}
	if s.Jar == nil || u == nil {
		return ret
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
		ctx := context.Background()
		var kookies []*kooky.Cookie
		if s.CookieStore != nil && len(s.cookies) == 0 {
			var err error
			kookies, err = s.CookieStore.TraverseCookies(s.filters...).ReadAllCookies(ctx)
			defer s.Close()
			if err != nil {
				s.initErr = err
				return
			}
		} else {
			kookies = kooky.FilterCookies(ctx, s.cookies, s.filters...).Collect(ctx)
		}
		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		if err != nil {
			s.initErr = err
			return
		}
		s.Jar = jar
		s.cookies = kookies
		cookies := kookies2cookies(ctx, kookies)
		setAllCookies(s, cookies)
	})

	return s.initErr
}

func (s *CookieJar) SubJar(ctx context.Context, filters ...kooky.Filter) (http.CookieJar, error) {
	if s == nil {
		return nil, errors.New(`nil receiver`)
	}
	if s.CookieStore == nil {
		return nil, errors.New(`no cookie store set`)
	}
	if err := s.InitJar(); err != nil {
		return nil, err
	}
	kookies := kooky.FilterCookies(ctx, s.cookies, filters...).Collect(ctx)
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}
	j := &CookieJar{
		filters:     append(s.filters, filters...),
		cookies:     kookies,
		Jar:         jar,
		CookieStore: s.CookieStore,
	}
	cookies := kookies2cookies(ctx, kookies)
	setAllCookies(jar, cookies)

	return j, nil
}

func kookies2cookies(ctx context.Context, kookies []*kooky.Cookie, filters ...kooky.Filter) []*http.Cookie {
	filteredKookies := kooky.FilterCookies(ctx, kookies, filters...).Collect(ctx)
	cookies := make([]*http.Cookie, 0, len(filteredKookies))
	for _, k := range filteredKookies {
		cookies = append(cookies, &k.Cookie)
	}
	return cookies
}

func setAllCookies(jar http.CookieJar, cookies []*http.Cookie) {
	cookieMap := make(map[url.URL][]*http.Cookie)

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
		u := url.URL{
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
		jar.SetCookies(&u, cs)
	}
}
