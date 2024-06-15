package kooky

import (
	"context"
	"errors"
	"iter"
	"net/http"
	"sync"
	"time"
)

// Cookie is the struct returned by functions in this package. Similar to http.Cookie.
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
	for cookie, _ := range s.OnlyCookies() {
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
