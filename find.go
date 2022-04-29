package kooky

import (
	"net/http"
	"sync"
)

// CookieStore represents a file, directory, etc containing cookies.
//
// Call CookieStore.Close() after using any of its methods.
type CookieStore interface {
	http.CookieJar
	SubJar(filters ...Filter) (http.CookieJar, error)
	ReadCookies(...Filter) ([]*Cookie, error)
	Browser() string
	Profile() string
	IsDefaultProfile() bool
	FilePath() string
	Close() error
}

// CookieStoreFinder tries to find cookie stores at default locations.
type CookieStoreFinder interface {
	FindCookieStores() ([]CookieStore, error)
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
//  import _ "github.com/zellyn/kooky/browser/all"
//
// Or only a specific browser:
//
//  import _ "github.com/zellyn/kooky/browser/chrome"
func FindAllCookieStores() []CookieStore {
	var ret []CookieStore

	var wg sync.WaitGroup
	wg.Add(len(finders))

	c := make(chan []CookieStore)
	done := make(chan struct{})

	go func() {
		for cookieStores := range c {
			ret = append(ret, cookieStores...)
		}
		close(done)
	}()

	muFinder.RLock()
	defer muFinder.RUnlock()
	for _, finder := range finders {
		go func(finder CookieStoreFinder) {
			defer wg.Done()
			cookieStores, err := finder.FindCookieStores()
			if err == nil && cookieStores != nil {
				c <- cookieStores
			}
		}(finder)
	}

	wg.Wait()
	close(c)

	<-done

	return ret
}

// ReadCookies() uses registered cookiestore finders to read cookies.
// Erronous reads are skipped.
//
// Register cookie store finders for all browsers like this:
//
//  import _ "github.com/zellyn/kooky/browser/all"
//
// Or only a specific browser:
//
//  import _ "github.com/zellyn/kooky/browser/chrome"
func ReadCookies(filters ...Filter) []*Cookie {
	var ret []*Cookie

	cs := make(chan []CookieStore)
	c := make(chan []*Cookie)
	done := make(chan struct{})

	// append cookies
	go func() {
		for cookies := range c {
			ret = append(ret, cookies...)
		}
		close(done)
	}()

	// read cookies
	go func() {
		var wgcs sync.WaitGroup
		for cookieStores := range cs {
			for _, store := range cookieStores {
				wgcs.Add(1)
				go func(store CookieStore) {
					defer wgcs.Done()
					cookies, err := store.ReadCookies(filters...)
					if err == nil && cookies != nil {
						c <- cookies
					}
				}(store)
			}

		}
		wgcs.Wait()
		close(c)
	}()

	// find cookie store
	var wgcsf sync.WaitGroup
	muFinder.RLock()
	defer muFinder.RUnlock()
	wgcsf.Add(len(finders))
	for _, finder := range finders {
		go func(finder CookieStoreFinder) {
			defer wgcsf.Done()
			cookieStores, err := finder.FindCookieStores()
			if err == nil && cookieStores != nil {
				cs <- cookieStores
			}
		}(finder)
	}
	wgcsf.Wait()
	close(cs)

	<-done

	return ret
}
