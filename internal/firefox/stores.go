package firefox

import (
	"iter"
	"path/filepath"
	"sync"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/firefox/find"
)

type storeOrErr struct {
	store kooky.CookieStore
	err   error
}

// CookieStoresForProfiles returns all cookie stores (SQLite + session) for
// the given profiles. Each profile is processed in its own goroutine to
// avoid blocking if a profile path is on a slow filesystem (e.g. network share).
// Errors from the profile iterator are forwarded to the caller.
func CookieStoresForProfiles(profiles iter.Seq2[find.Profile, error]) kooky.CookieStoreSeq {
	if profiles == nil {
		return func(yield func(kooky.CookieStore, error) bool) {}
	}
	return func(yield func(kooky.CookieStore, error) bool) {
		ch := make(chan storeOrErr)
		quit := make(chan struct{})

		var wg sync.WaitGroup
		// feed goroutine: reads profiles, spawns per-profile goroutines
		wg.Add(1)
		go func() {
			defer wg.Done()
			var inner sync.WaitGroup
			for p, err := range profiles {
				if err != nil {
					select {
					case ch <- storeOrErr{err: err}:
					case <-quit:
						return
					}
					continue
				}
				inner.Add(1)
				go func(p find.Profile) {
					defer inner.Done()
					for _, st := range cookieStoresForProfile(p) {
						select {
						case ch <- storeOrErr{store: st}:
						case <-quit:
							return
						}
					}
				}(p)
			}
			inner.Wait()
		}()
		go func() {
			wg.Wait()
			close(ch)
		}()

		for item := range ch {
			if !yield(item.store, item.err) {
				close(quit)
				for range ch {
				}
				return
			}
		}
	}
}

// cookieStoresForProfile returns the SQLite and session cookie stores for a single profile.
func cookieStoresForProfile(p find.Profile) []*cookies.CookieJar {
	return []*cookies.CookieJar{
		{
			CookieStore: &CookieStore{
				DefaultCookieStore: cookies.DefaultCookieStore{
					BrowserStr:           p.Browser,
					ProfileStr:           p.Name,
					IsDefaultProfileBool: p.IsDefaultProfile,
					FileNameStr:          filepath.Join(p.Path, `cookies.sqlite`),
				},
			},
		},
		{
			CookieStore: &SessionCookieStore{
				DefaultCookieStore: cookies.DefaultCookieStore{
					BrowserStr:           p.Browser,
					ProfileStr:           p.Name,
					IsDefaultProfileBool: p.IsDefaultProfile,
					FileNameStr:          p.Path,
				},
			},
		},
	}
}
