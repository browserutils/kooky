package kooky

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// Filter is used for filtering cookies in ReadCookies() functions.
// Filter order might be changed for performance reasons
// (omission of value decryption of filtered out cookies, etc).
//
// A cookie passes the Filter if Filter.Filter returns true.
type Filter interface{ Filter(*Cookie) bool }

type FilterFunc func(*Cookie) bool

func (f FilterFunc) Filter(c *Cookie) bool {
	if f == nil {
		return false
	}
	return f(c)
}

type ValueFilterFunc func(*Cookie) bool

func (f ValueFilterFunc) Filter(c *Cookie) bool {
	if f == nil {
		return false
	}
	return f(c)
}

func FilterCookies[S CookieSeq | ~[]*Cookie | ~[]*http.Cookie](ctx context.Context, cookies S, filters ...Filter) CookieSeq {
	var ret CookieSeq
	// https://github.com/golang/go/issues/45380#issuecomment-1014950980
	switch cookiesTyped := any(cookies).(type) {
	case CookieSeq:
		ret = filterCookieSeq(ctx, cookiesTyped, filters...)
	case Cookies:
		ret = filterCookieSlice(ctx, []*Cookie(cookiesTyped), filters...)
	case []*Cookie:
		ret = filterCookieSlice(ctx, cookiesTyped, filters...)
	case []*http.Cookie:
		ret = filterCookieSlice(ctx, cookiesTyped, filters...)
	default:
		rv := reflect.ValueOf(cookies)
		rtc := reflect.TypeFor[[]*Cookie]()
		rthc := reflect.TypeFor[[]*http.Cookie]()
		if rv.CanConvert(rtc) {
			cookiesTyped := rv.Convert(rtc).Interface().([]*Cookie)
			ret = filterCookieSlice(ctx, cookiesTyped, filters...)
		} else if rv.CanConvert(rthc) {
			cookiesTyped := rv.Convert(rthc).Interface().([]*http.Cookie)
			ret = filterCookieSlice(ctx, cookiesTyped, filters...)
		} else {
			ret = func(yield func(*Cookie, error) bool) { yield(nil, errors.New(`unknown type`)) }
		}
	}
	return ret
}

func filterCookieSeq(ctx context.Context, cookies CookieSeq, filters ...Filter) CookieSeq {
	return func(yield func(*Cookie, error) bool) {
		if cookies == nil {
			yield(nil, errors.New(`nil receiver`))
			return
		}
		for cookie, errCookie := range cookies {
			if errCookie != nil {
				if !yield(nil, errCookie) {
					return
				}
				continue
			}
			if cookie == nil {
				continue
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
			if !FilterCookie(ctx, cookie, filters...) {
				continue
			}
			if !yield(cookie, nil) {
				return
			}
		}
	}
}

func filterCookieSlice[S ~[]*T, T Cookie | http.Cookie](ctx context.Context, cookies S, filters ...Filter) CookieSeq {
	return func(yield func(*Cookie, error) bool) {
		if len(cookies) < 1 {
			_ = yield(nil, errors.New(`cookie slice of lenght 0`))
			return
		}
		switch cookiesTyped := any(cookies).(type) {
		case []*http.Cookie:
		cookieLoopHTTP:
			for _, cookie := range cookiesTyped {
				if cookie == nil {
					continue
				}
				select {
				case <-ctx.Done():
					break cookieLoopHTTP
				default:
				}
				kooky := &Cookie{Cookie: *cookie}
				if !FilterCookie(ctx, kooky, filters...) {
					continue
				}
				if !yield(kooky, nil) {
					return
				}
			}
		case []*Cookie:
		cookieLoopKooky:
			for _, cookie := range cookiesTyped {
				if cookie == nil {
					continue
				}
				select {
				case <-ctx.Done():
					break cookieLoopKooky
				default:
				}
				if !FilterCookie(ctx, cookie, filters...) {
					continue
				}
				if !yield(cookie, nil) {
					return
				}
			}
		}
	}
}

// FilterCookie() tells if a "cookie" passes all "filters".
func FilterCookie[T Cookie | http.Cookie](ctx context.Context, cookie *T, filters ...Filter) bool {
	if cookie == nil {
		return false
	}

	var c *Cookie
	// https://github.com/golang/go/issues/45380#issuecomment-1014950980
	switch cookieTyp := any(cookie).(type) {
	case *http.Cookie:
		c = &Cookie{Cookie: *cookieTyp}
	case *Cookie:
		c = cookieTyp
	}

	for _, filter := range filters {
		if filter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			return false
		default:
		}
		if !filter.Filter(c) {
			return false
		}
	}
	return true
}

// debug filter

// Debug prints the cookie.
//
// Position Debug after the filter you want to test.
var Debug Filter = FilterFunc(func(cookie *Cookie) bool {
	// TODO(srlehn): where should the Debug filter be positioned when the filter rearrangement happens?
	fmt.Printf("%+#v\n", cookie)
	return true
})

// domain filters

type domainFilter struct {
	filterFunc FilterFunc
	typ        string
	domain     string
}

func (d *domainFilter) Type() string {
	if d == nil {
		return ``
	}
	return d.typ
}
func (d *domainFilter) Domain() string {
	if d == nil {
		return ``
	}
	return d.domain
}
func (d *domainFilter) Filter(c *Cookie) bool {
	return d != nil && d.filterFunc != nil && d.filterFunc(c)
}

func Domain(domain string) Filter {
	f := func(cookie *Cookie) bool {
		return cookie != nil && cookie.Domain == domain
	}
	return &domainFilter{filterFunc: f, typ: `domain`, domain: domain}
}
func DomainContains(substr string) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && strings.Contains(cookie.Domain, substr)
	})
}
func DomainHasPrefix(prefix string) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && strings.HasPrefix(cookie.Domain, prefix)
	})
}
func DomainHasSuffix(suffix string) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && strings.HasSuffix(cookie.Domain, suffix)
	})
}

// name filters

func Name(name string) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && cookie.Name == name
	})
}
func NameContains(substr string) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && strings.Contains(cookie.Name, substr)
	})
}
func NameHasPrefix(prefix string) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && strings.HasPrefix(cookie.Name, prefix)
	})
}
func NameHasSuffix(suffix string) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && strings.HasSuffix(cookie.Name, suffix)
	})
}

// path filters

func Path(path string) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && cookie.Path == path
	})
}
func PathContains(substr string) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && strings.Contains(cookie.Path, substr)
	})
}
func PathHasPrefix(prefix string) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && strings.HasPrefix(cookie.Path, prefix)
	})
}
func PathHasSuffix(suffix string) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && strings.HasSuffix(cookie.Path, suffix)
	})
}
func PathDepth(depth int) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && strings.Count(strings.TrimRight(cookie.Path, `/`), `/`) == depth
	})
}

// value filters

func Value(value string) Filter {
	return ValueFilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && cookie.Value == value
	})
}
func ValueContains(substr string) Filter {
	return ValueFilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && strings.Contains(cookie.Value, substr)
	})
}
func ValueHasPrefix(prefix string) Filter {
	return ValueFilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && strings.HasPrefix(cookie.Value, prefix)
	})
}
func ValueHasSuffix(suffix string) Filter {
	return ValueFilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && strings.HasSuffix(cookie.Value, suffix)
	})
}
func ValueLen(length int) Filter {
	return ValueFilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && len(cookie.Value) == length
	})
}

// secure filter

var Secure Filter = FilterFunc(func(cookie *Cookie) bool {
	return cookie != nil && cookie.Secure
})

// httpOnly filter

var HTTPOnly Filter = FilterFunc(func(cookie *Cookie) bool {
	return cookie != nil && cookie.HttpOnly
})

// expires filters

var Valid Filter = ValueFilterFunc(func(cookie *Cookie) bool {
	return cookie != nil && cookie.Expires.After(time.Now()) && cookie.Cookie.Valid() == nil
})

var Expired Filter = FilterFunc(func(cookie *Cookie) bool {
	return cookie != nil && cookie.Expires.Before(time.Now())
})

func ExpiresAfter(u time.Time) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && cookie.Expires.After(u)
	})
}
func ExpiresBefore(u time.Time) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && cookie.Expires.Before(u)
	})
}

// creation filters

func CreationAfter(u time.Time) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && cookie.Creation.After(u)
	})
}
func CreationBefore(u time.Time) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && cookie.Creation.Before(u)
	})
}
