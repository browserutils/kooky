package kooky

import (
	"context"
	"iter"
	"net/http"
	"runtime"
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
	// FindCookieStores() ([]CookieStore, error)
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

func (s CookieStoreSeq) AllCookieStores(ctx context.Context) []CookieStore {
	var ret []CookieStore
	if s == nil {
		return nil
	}
Outer:
	for cookieStore, _ := range s {
		select {
		case <-ctx.Done():
			break Outer
		default:
		}
		ret = append(ret, cookieStore)
	}
	return ret
}

func (s CookieStoreSeq) TraverseCookies(ctx context.Context, filters ...Filter) CookieSeq {
	return func(yield func(*Cookie, error) bool) {
		if s == nil {
			return
		}
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
			for cookieStore, _ := range s {
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

func TraverseCookieStores(ctx context.Context) CookieStoreSeq {
	return func(yield func(CookieStore, error) bool) {
		var wg sync.WaitGroup
		wg.Add(len(finders))

		c := make(chan CookieStore)
		done := make(chan struct{})

		go func() {
			for cookieStore := range c {
				select {
				case <-ctx.Done():
					return
				default:
				}
				if !yield(cookieStore, nil) {
					return
				}

			}
			close(done)
		}()

		muFinder.RLock()
		defer muFinder.RUnlock()
		for _, finder := range finders {
			select {
			case <-ctx.Done():
				return
			default:
			}
			go func(finder CookieStoreFinder) {
				defer wg.Done()
				for cookieStore, err := range finder.FindCookieStores() {
					if err == nil && cookieStore != nil {
						c <- cookieStore
					}
				}
			}(finder)
		}

		wg.Wait()
		close(c)

		<-done
	}
}
