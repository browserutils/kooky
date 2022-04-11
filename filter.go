package kooky

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Filter is used for filtering cokies in ReadCookies() functions.
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

// FilterCookies() applies "filters" in order to the "cookies".
func FilterCookies[T Cookie | http.Cookie](cookies []*T, filters ...Filter) []*T {
	var ret = make([]*T, 0, len(cookies))

	// https://github.com/golang/go/issues/45380#issuecomment-1014950980
	switch cookiesTyp := any(cookies).(type) {
	case []*http.Cookie:
	cookieLoopHTTP:
		for i, cookie := range cookiesTyp {
			if cookie == nil {
				continue
			}
			for _, filter := range filters {
				if !filter.Filter(&Cookie{Cookie: *cookie}) {
					continue cookieLoopHTTP
				}
			}
			ret = append(ret, cookies[i])
		}
	case []*Cookie:
	cookieLoopKooky:
		for i, cookie := range cookiesTyp {
			if cookie == nil {
				continue
			}
			for _, filter := range filters {
				if !filter.Filter(cookie) {
					continue cookieLoopKooky
				}
			}
			ret = append(ret, cookies[i])
		}
	}

	return ret
}

// FilterCookie() tells if a "cookie" passes all "filters".
func FilterCookie[T Cookie | http.Cookie](cookie *T, filters ...Filter) bool {
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
	fmt.Printf("%+#v\n", cookie)
	return true
})

// domain filters

type domainFilter struct {
	filterFunc func(*Cookie) bool
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
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && cookie.Value == value
	})
}
func ValueContains(substr string) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && strings.Contains(cookie.Value, substr)
	})
}
func ValueHasPrefix(prefix string) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && strings.HasPrefix(cookie.Value, prefix)
	})
}
func ValueHasSuffix(suffix string) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
		return cookie != nil && strings.HasSuffix(cookie.Value, suffix)
	})
}
func ValueLen(length int) Filter {
	return FilterFunc(func(cookie *Cookie) bool {
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

var Valid Filter = FilterFunc(func(cookie *Cookie) bool {
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
