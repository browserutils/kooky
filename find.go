package kooky

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"sync"
)

// CookieStore represents a file, directory, etc containing cookies.
//
// Call CookieStore.Close() after using any of its methods.
type CookieStore interface {
	http.CookieJar
	BrowserInfo
	SubJar(context.Context, ...Filter) (http.CookieJar, error)
	TraverseCookies(...Filter) CookieSeq
	Close() error
}

type BrowserInfo interface {
	Browser() string
	Profile() string
	IsDefaultProfile() bool
	FilePath() string
}

// CookieStoreFinder tries to find cookie stores at default locations.
type CookieStoreFinder interface {
	FindCookieStores() CookieStoreSeq
}

var (
	finders  = map[string]CookieStoreFinder{}
	muFinder sync.RWMutex
)

// RegisterFinder() registers CookieStoreFinder enabling automatic finding of
// cookie stores with FindAllCookieStores() and ReadCookies().
//
// RegisterFinder() is called by init() in the browser subdirectories.
func RegisterFinder(browser string, finder CookieStoreFinder) {
	muFinder.Lock()
	defer muFinder.Unlock()
	if finder != nil {
		finders[browser] = finder
	}
}

// FindAllCookieStores() tries to find cookie stores at default locations.
//
// FindAllCookieStores() requires registered CookieStoreFinders.
//
// Register cookie store finders for all browsers like this:
//
//	import _ "github.com/browserutils/kooky/browser/all"
//
// Or only a specific browser:
//
//	import _ "github.com/browserutils/kooky/browser/chrome"
func FindAllCookieStores(ctx context.Context) []CookieStore {
	return TraverseCookieStores(ctx).AllCookieStores(ctx)
}

type CookieStoreSeq iter.Seq2[CookieStore, error]

// sequence of non-nil cookie stores and nil errors
func (s CookieStoreSeq) OnlyCookieStores() CookieStoreSeq {
	return func(yield func(CookieStore, error) bool) {
		if s == nil {
			return
		}
		for cookieStore, err := range s {
			if err != nil || cookieStore == nil {
				continue
			}
			if !yield(cookieStore, nil) {
				return
			}
		}
	}
}

func (s CookieStoreSeq) AllCookieStores(ctx context.Context) []CookieStore {
	var ret []CookieStore
	if s == nil {
		return nil
	}
Outer:
	for cookieStore := range s {
		select {
		case <-ctx.Done():
			break Outer
		default:
		}
		if cookieStore == nil {
			continue
		}
		ret = append(ret, cookieStore)
	}
	return ret
}

func (s CookieStoreSeq) TraverseCookies(ctx context.Context, filters ...Filter) CookieSeq {
	if s == nil {
		return func(yield func(*Cookie, error) bool) {}
	}

	ctx, cancel := context.WithCancel(ctx)
	type ce struct {
		c *Cookie
		e error
	}
	startChan := make(chan struct{}, 1)
	cookieChan := make(chan ce, 1)

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-startChan: // wait for iteration start
		}
		var wg sync.WaitGroup
		defer func() {
			wg.Wait()
			cancel()
			close(cookieChan)
		}()
		for cookieStore, err := range s {
			if err != nil {
				select {
				case <-ctx.Done():
					return
				case cookieChan <- ce{e: fmt.Errorf(`cookie store: %w`, err)}:
				}
				continue
			}
			if cookieStore == nil {
				continue
			}
			wg.Add(1)
			go func(cookieStore CookieStore) {
				defer wg.Done()
				for cookie, err := range cookieStore.TraverseCookies(filters...) {
					select {
					case <-ctx.Done():
						return
					case cookieChan <- ce{c: cookie, e: err}:
					}
				}
			}(cookieStore)
		}
	}()

	return func(yield func(*Cookie, error) bool) {
		startChan <- struct{}{}
		for {
			select {
			case <-ctx.Done():
				return
			case c, ok := <-cookieChan:
				if !ok {
					cancel()
					return
				}
				if !yield(c.c, c.e) {
					cancel()
					return
				}
			}
		}
	}
}

func TraverseCookieStores(ctx context.Context) CookieStoreSeq {
	ctx, cancel := context.WithCancel(ctx)
	type se struct {
		s CookieStore
		e error
	}
	startChan := make(chan struct{}, 1)
	storeChan := make(chan se, 1)

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-startChan:
		}

		var wg sync.WaitGroup
		wg.Add(len(finders))
		defer func() {
			wg.Wait()
			cancel()
			close(storeChan)
		}()

		muFinder.RLock()
		defer muFinder.RUnlock()

		for _, finder := range finders {
			if finder == nil {
				wg.Done()
				continue
			}
			go func(finder CookieStoreFinder) {
				defer wg.Done()
				for cookieStore, err := range finder.FindCookieStores() {
					select {
					case <-ctx.Done():
						return
					case storeChan <- se{s: cookieStore, e: err}:
					}
				}
			}(finder)
		}
	}()

	return func(yield func(CookieStore, error) bool) {
		startChan <- struct{}{}
		for {
			select {
			case <-ctx.Done():
				return
			case s, ok := <-storeChan:
				if !ok {
					cancel()
					return
				}
				if !yield(s.s, s.e) {
					cancel()
					return
				}
			}
		}
	}
}
