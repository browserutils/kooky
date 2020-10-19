package kooky

import (
	"fmt"
	"strings"
	"time"
)

// passed cookies are non-nil
type Filter func(*Cookie) bool

func FilterCookies(cookies []*Cookie, filters ...Filter) []*Cookie {
	var ret = make([]*Cookie, 0, len(cookies))
cookieLoop:
	for _, cookie := range cookies {
		if cookie == nil {
			continue
		}
		for _, filter := range filters {
			if !filter(cookie) {
				continue cookieLoop
			}
		}
		ret = append(ret, cookie)
	}
	return ret
}

func FilterCookie(cookie *Cookie, filters ...Filter) bool {
	if cookie == nil {
		return false
	}
	for _, filter := range filters {
		if !filter(cookie) {
			return false
		}
	}
	return true
}

// debug filter

var Debug Filter = func(cookie *Cookie) bool {
	fmt.Printf("%+#v\n", cookie)
	return cookie != nil && true
}

// domain filters

func Domain(domain string) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && cookie.Domain == domain
	}
}
func DomainContains(substr string) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && strings.Contains(cookie.Domain, substr)
	}
}
func DomainHasPrefix(prefix string) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && strings.HasPrefix(cookie.Domain, prefix)
	}
}
func DomainHasSuffix(suffix string) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && strings.HasSuffix(cookie.Domain, suffix)
	}
}

// name filters

func Name(name string) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && cookie.Name == name
	}
}
func NameContains(substr string) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && strings.Contains(cookie.Name, substr)
	}
}
func NameHasPrefix(prefix string) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && strings.HasPrefix(cookie.Name, prefix)
	}
}
func NameHasSuffix(suffix string) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && strings.HasSuffix(cookie.Name, suffix)
	}
}

// path filters

func Path(path string) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && cookie.Path == path
	}
}
func PathContains(substr string) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && strings.Contains(cookie.Path, substr)
	}
}
func PathHasPrefix(prefix string) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && strings.HasPrefix(cookie.Path, prefix)
	}
}
func PathHasSuffix(suffix string) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && strings.HasSuffix(cookie.Path, suffix)
	}
}
func PathDepth(depth int) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && strings.Count(strings.TrimRight(cookie.Path, `/`), `/`) == depth
	}
}

// value filters

func Value(value string) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && cookie.Value == value
	}
}
func ValueContains(substr string) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && strings.Contains(cookie.Value, substr)
	}
}
func ValueHasPrefix(prefix string) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && strings.HasPrefix(cookie.Value, prefix)
	}
}
func ValueHasSuffix(suffix string) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && strings.HasSuffix(cookie.Value, suffix)
	}
}
func ValueLen(length int) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && len(cookie.Value) == length
	}
}

// secure filter

var Secure Filter = func(cookie *Cookie) bool {
	return cookie != nil && cookie.Secure
}

// httpOnly filter

var HTTPOnly Filter = func(cookie *Cookie) bool {
	return cookie != nil && cookie.HttpOnly
}

// expires filters

func ExpiresAfter(u time.Time) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && cookie.Expires.After(u)
	}
}
func ExpiresBefore(u time.Time) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && cookie.Expires.Before(u)
	}
}

// creation filters

func CreationAfter(u time.Time) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && cookie.Creation.After(u)
	}
}
func CreationBefore(u time.Time) Filter {
	return func(cookie *Cookie) bool {
		return cookie != nil && cookie.Creation.Before(u)
	}
}
