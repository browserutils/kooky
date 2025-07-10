package kooky

import (
	"context"
	"encoding/json"
	"errors"
	"iter"
	"net/http"
	"sync"
	"time"
)

// Cookie is an http.Cookie augmented with information obtained through the scraping process.
type Cookie struct {
	http.Cookie
	Creation  time.Time
	Container string
	Browser   BrowserInfo
}

// Cookie retrieving functions in this package like TraverseCookies(), ReadCookies(), AllCookies()
// use registered cookiestore finders to read cookies.
// Erronous reads are skipped.
//
// Register cookie store finders for all browsers like this:
//
//	import _ "github.com/browserutils/kooky/browser/all"
//
// Or only a specific browser:
//
//	import _ "github.com/browserutils/kooky/browser/chrome"
func ReadCookies(ctx context.Context, filters ...Filter) (Cookies, error) {
	return TraverseCookies(ctx).ReadAllCookies(ctx)
}

func AllCookies(filters ...Filter) Cookies {
	// for convenience...
	ctx := context.Background()
	return TraverseCookies(ctx).Collect(ctx)
}

// adjustments to the json marshaling to allow dates with more than 4 year digits
// https://github.com/golang/go/issues/4556
// https://github.com/golang/go/issues/54580
// encoding/json/v2 "format"(?) might make this unnecessary

func (c *Cookie) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte(`null`), nil
	}
	c2 := &struct {
		// net/http.Cookie
		Name        string        `json:"name"`
		Value       string        `json:"value"`
		Quoted      bool          `json:"quoted"`
		Path        string        `json:"path"`
		Domain      string        `json:"domain"`
		Expires     jsonTime      `json:"expires"`
		RawExpires  string        `json:"raw_expires,omitempty"`
		MaxAge      int           `json:"max_age"`
		Secure      bool          `json:"secure"`
		HttpOnly    bool          `json:"http_only"`
		SameSite    http.SameSite `json:"same_site"`
		Partitioned bool          `json:"partitioned"`
		Raw         string        `json:"raw,omitempty"`
		Unparsed    []string      `json:"unparsed,omitempty"`
		// extra fields
		Creation         jsonTime `json:"creation"`
		Browser          string   `json:"browser,omitempty"`
		Profile          string   `json:"profile,omitempty"`
		IsDefaultProfile bool     `json:"is_default_profile"`
		Container        string   `json:"container,omitempty"`
		FilePath         string   `json:"file_path,omitempty"`
	}{
		Name:        c.Cookie.Name,
		Value:       c.Cookie.Value,
		Quoted:      c.Cookie.Quoted,
		Path:        c.Cookie.Path,
		Domain:      c.Cookie.Domain,
		Expires:     jsonTime{c.Cookie.Expires},
		RawExpires:  c.Cookie.RawExpires,
		MaxAge:      c.Cookie.MaxAge,
		Secure:      c.Cookie.Secure,
		HttpOnly:    c.Cookie.HttpOnly,
		SameSite:    c.Cookie.SameSite,
		Partitioned: c.Cookie.Partitioned,
		Raw:         c.Cookie.Raw,
		Unparsed:    c.Cookie.Unparsed,
		Creation:    jsonTime{c.Creation},
		Container:   c.Container,
	}
	if c.Browser != nil {
		c2.Browser = c.Browser.Browser()
		c2.Profile = c.Browser.Profile()
		c2.IsDefaultProfile = c.Browser.IsDefaultProfile()
		c2.FilePath = c.Browser.FilePath()
	}
	return json.Marshal(c2)
}

type jsonTime struct{ time.Time }

// MarshalJSON implements the [json.Marshaler] interface.
// The time is a quoted string in the RFC 3339 format with sub-second precision.
// the timestamp might be represented as invalid RFC 3339 if necessary (year with more than 4 digits).
func (t jsonTime) MarshalJSON() ([]byte, error) {
	if b, err := t.Time.MarshalJSON(); err == nil {
		return b, nil
	}
	return []byte(t.Time.Format(`"` + time.RFC3339 + `"`)), nil
}

// for-rangeable cookie retriever
type CookieSeq iter.Seq2[*Cookie, error]

func TraverseCookies(ctx context.Context, filters ...Filter) CookieSeq {
	return TraverseCookieStores(ctx).TraverseCookies(ctx, filters...)
}

// Collect() is the same as ReadAllCookies but ignores the error
func (s CookieSeq) Collect(ctx context.Context) Cookies {
	cookies, _ := s.ReadAllCookies(ctx)
	return cookies
}

func (s CookieSeq) ReadAllCookies(ctx context.Context) (Cookies, error) {
	if s == nil {
		return nil, errors.New(`nil receiver`)
	}
	var (
		errs    []error
		cookies []*Cookie
	)
Outer:
	for cookie, err := range s {
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if cookie != nil {
			cookies = append(cookies, cookie)
		}
		select {
		case <-ctx.Done():
			errs = append(errs, errors.New(`context cancel`))
			break Outer
		default:
		}
	}
	return cookies, errors.Join(errs...)
}

// sequence of non-nil cookies and nil errors
func (s CookieSeq) OnlyCookies() CookieSeq {
	return func(yield func(*Cookie, error) bool) {
		if s == nil {
			return
		}
		for cookie, err := range s {
			if err != nil || cookie == nil {
				continue
			}
			if !yield(cookie, nil) {
				return
			}
		}
	}
}

func (s CookieSeq) Filter(ctx context.Context, filters ...Filter) CookieSeq {
	return func(yield func(*Cookie, error) bool) {
		if s == nil {
			yield(nil, errors.New(`nil receiver`))
			return
		}
		for cookie, errCookie := range s {
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

func (s CookieSeq) FirstMatch(ctx context.Context, filters ...Filter) *Cookie {
	if s == nil {
		return nil
	}
	for cookie := range s.OnlyCookies() {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		if FilterCookie(ctx, cookie, filters...) {
			return cookie
		}
	}
	return nil
}

func (s CookieSeq) Merge(seqs ...CookieSeq) CookieSeq { return MergeCookieSeqs(append(seqs, s)...) }

func MergeCookieSeqs(seqs ...CookieSeq) CookieSeq {
	var sq []iter.Seq2[*Cookie, error]
	for _, s := range seqs {
		sq = append(sq, iter.Seq2[*Cookie, error](s))
	}
	return CookieSeq(mergeSeqs(sq...))
}

func mergeSeqs[S iter.Seq2[T, error], T any](seqs ...S) S {
	seqs0 := func(yield func(T, error) bool) {}
	seqs2 := func(yield func(T, error) bool) {
		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(len(seqs) + 1)
		runner := func(seq S) {
			defer wg.Done()
			if seq == nil {
				return
			}
			for v, error := range seq {
				if !yield(v, error) {
					return
				}
			}
		}
		for _, seq := range seqs {
			go runner(seq)
		}
	}
	switch len(seqs) {
	case 0:
		return seqs0
	case 1:
		return seqs[0]
	default:
		return seqs2
	}
}

func (s CookieSeq) Chan(ctx context.Context) <-chan *Cookie {
	cookieChan := make(chan *Cookie)
	go func() {
		defer close(cookieChan)
		for cookie, err := range s {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err != nil || cookie == nil {
				continue
			}
			cookieChan <- cookie
		}
	}()
	return cookieChan
}

type Cookies []*Cookie

func (c Cookies) Seq() CookieSeq {
	return func(yield func(*Cookie, error) bool) {
		if c == nil {
			return
		}
		for _, cookie := range c {
			if cookie == nil {
				continue
			}
			if !yield(cookie, nil) {
				return
			}
		}
	}
}
