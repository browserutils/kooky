package filter

import (
	"fmt"
	"strings"
	"time"

	"github.com/browserutils/kooky"
)

// FilterFunc is the simplest way to implement a [kooky.Filter].
type FilterFunc func(*kooky.Cookie) bool

func (f FilterFunc) Filter(c *kooky.Cookie) bool {
	if f == nil {
		return false
	}
	return f(c)
}

// ValueFilterFunc marks a [kooky.Filter] as depending on the cookie value.
// This allows filter-aware cookie stores to defer expensive operations like
// decryption until non-value filters have passed.
type ValueFilterFunc func(*kooky.Cookie) bool

func (f ValueFilterFunc) Filter(c *kooky.Cookie) bool {
	if f == nil {
		return false
	}
	return f(c)
}

// debug filter

// Debug prints the cookie.
//
// Position Debug after the filter you want to test.
var Debug kooky.Filter = FilterFunc(func(cookie *kooky.Cookie) bool {
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
func (d *domainFilter) Filter(c *kooky.Cookie) bool {
	return d != nil && d.filterFunc != nil && d.filterFunc(c)
}

func Domain(domain string) kooky.Filter {
	f := func(cookie *kooky.Cookie) bool {
		return cookie != nil && cookie.Domain == domain
	}
	return &domainFilter{filterFunc: f, typ: `domain`, domain: domain}
}
func DomainContains(substr string) kooky.Filter {
	return FilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && strings.Contains(cookie.Domain, substr)
	})
}
func DomainHasPrefix(prefix string) kooky.Filter {
	return FilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && strings.HasPrefix(cookie.Domain, prefix)
	})
}
func DomainHasSuffix(suffix string) kooky.Filter {
	return FilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && strings.HasSuffix(cookie.Domain, suffix)
	})
}

// name filters

func Name(name string) kooky.Filter {
	return FilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && cookie.Name == name
	})
}
func NameContains(substr string) kooky.Filter {
	return FilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && strings.Contains(cookie.Name, substr)
	})
}
func NameHasPrefix(prefix string) kooky.Filter {
	return FilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && strings.HasPrefix(cookie.Name, prefix)
	})
}
func NameHasSuffix(suffix string) kooky.Filter {
	return FilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && strings.HasSuffix(cookie.Name, suffix)
	})
}

// path filters

func Path(path string) kooky.Filter {
	return FilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && cookie.Path == path
	})
}
func PathContains(substr string) kooky.Filter {
	return FilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && strings.Contains(cookie.Path, substr)
	})
}
func PathHasPrefix(prefix string) kooky.Filter {
	return FilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && strings.HasPrefix(cookie.Path, prefix)
	})
}
func PathHasSuffix(suffix string) kooky.Filter {
	return FilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && strings.HasSuffix(cookie.Path, suffix)
	})
}
func PathDepth(depth int) kooky.Filter {
	return FilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && strings.Count(strings.TrimRight(cookie.Path, `/`), `/`) == depth
	})
}

// value filters

func Value(value string) kooky.Filter {
	return ValueFilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && cookie.Value == value
	})
}
func ValueContains(substr string) kooky.Filter {
	return ValueFilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && strings.Contains(cookie.Value, substr)
	})
}
func ValueHasPrefix(prefix string) kooky.Filter {
	return ValueFilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && strings.HasPrefix(cookie.Value, prefix)
	})
}
func ValueHasSuffix(suffix string) kooky.Filter {
	return ValueFilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && strings.HasSuffix(cookie.Value, suffix)
	})
}
func ValueLen(length int) kooky.Filter {
	return ValueFilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && len(cookie.Value) == length
	})
}

// secure filter

var Secure kooky.Filter = FilterFunc(func(cookie *kooky.Cookie) bool {
	return cookie != nil && cookie.Secure
})

// httpOnly filter

var HTTPOnly kooky.Filter = FilterFunc(func(cookie *kooky.Cookie) bool {
	return cookie != nil && cookie.HttpOnly
})

// expires filters

var Valid kooky.Filter = ValueFilterFunc(func(cookie *kooky.Cookie) bool {
	return cookie != nil && cookie.Expires.After(time.Now()) && cookie.Cookie.Valid() == nil
})

var Expired kooky.Filter = FilterFunc(func(cookie *kooky.Cookie) bool {
	return cookie != nil && cookie.Expires.Before(time.Now())
})

func ExpiresAfter(u time.Time) kooky.Filter {
	return FilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && cookie.Expires.After(u)
	})
}
func ExpiresBefore(u time.Time) kooky.Filter {
	return FilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && cookie.Expires.Before(u)
	})
}

// creation filters

func CreationAfter(u time.Time) kooky.Filter {
	return FilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && cookie.Creation.After(u)
	})
}
func CreationBefore(u time.Time) kooky.Filter {
	return FilterFunc(func(cookie *kooky.Cookie) bool {
		return cookie != nil && cookie.Creation.Before(u)
	})
}
