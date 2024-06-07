package kooky

import (
	"context"
	"errors"
	"iter"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// TODO(zellyn): figure out what to do with quoted values, like the "bcookie" cookie
// from slideshare.net

// Cookie is the struct returned by functions in this package. Similar to http.Cookie.
type Cookie struct {
	http.Cookie
	Creation  time.Time
	Container string
	Browser   BrowserInfo
}

// ReadCookies() uses registered cookiestore finders to read cookies.
// Erronous reads are skipped.
//
// Register cookie store finders for all browsers like this:
//
//	import _ "github.com/browserutils/kooky/browser/all"
//
// Or only a specific browser:
//
//	import _ "github.com/browserutils/kooky/browser/chrome"
func ReadCookies(filters ...Filter) []*Cookie {
	var cookies []*Cookie
	for cookie, err := range TraverseCookies(context.Background()) {
		if err != nil || cookie == nil {
			continue
		}
		cookies = append(cookies, cookie)
	}
	return cookies
}

type CookieSeq iter.Seq2[*Cookie, error]

func TraverseCookies(ctx context.Context, filters ...Filter) CookieSeq {
	return func(yield func(*Cookie, error) bool) {
		ctx, cancel := context.WithCancel(ctx)
		type ce struct {
			c *Cookie
			e error
		}
		cookieChan := make(chan ce)

		var wgTot sync.WaitGroup
		defer wgTot.Wait()
		wgTot.Add(1)
		go func() {
			defer wgTot.Done()

			var wgTrav sync.WaitGroup
			defer func() {
				wgTrav.Wait()
				cancel()
				close(cookieChan)
			}()
			for cookieStore, _ := range TraverseCookieStores(ctx) {
				select {
				case <-ctx.Done():
					return
				default:
				}
				wgTrav.Add(1)
				go func(cookieStore CookieStore) {
					defer wgTrav.Done()
					for cookie, err := range cookieStore.TraverseCookies(filters...) {
						select {
						case <-ctx.Done():
							return
						default:
						}
						cookieChan <- ce{c: cookie, e: err}
					}
				}(cookieStore)
			}
		}()

		wgTot.Add(runtime.NumCPU())
		for range runtime.NumCPU() {
			go func() {
				defer wgTot.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case c, ok := <-cookieChan:
						if !ok {
							return
						}
						if !yield(c.c, c.e) {
							cancel()
							return
						}
					}
				}
			}()
		}
	}
}

func (s CookieSeq) ReadAllCookies(ctx context.Context) ([]*Cookie, error) {
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

func (s CookieSeq) OnlyCookies() iter.Seq[*Cookie] {
	return func(yield func(*Cookie) bool) {
		if s == nil {
			return
		}
		for cookie, err := range s {
			if err != nil || cookie == nil {
				continue
			}
			if !yield(cookie) {
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
