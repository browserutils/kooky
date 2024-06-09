//go:build !windows

package konqueror

import (
	"os"
	"path/filepath"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
)

type konquerorFinder struct{}

var _ kooky.CookieStoreFinder = (*konquerorFinder)(nil)

func init() {
	kooky.RegisterFinder(`konqueror`, &konquerorFinder{})
}

func (f *konquerorFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		//var stInner *cookies.DefaultCookieStore
		/*defer func() {
			if stInner == nil {
				return
			}
			stInner.IsDefaultProfileBool = true
		}()*/
		for root, err := range konquerorRoots {
			if err != nil {
				if !yield(nil, err) {
					return
				}
			}
			stInner := &cookies.DefaultCookieStore{
				BrowserStr:  `konqueror`,
				FileNameStr: filepath.Join(root, `kcookiejar`, `cookies`),
			}
			st := &cookies.CookieJar{CookieStore: &konquerorCookieStore{DefaultCookieStore: *stInner}}
			if !yield(st, nil) {
				return
			}
		}
	}
}

func konquerorRoots(yield func(string, error) bool) {
	// fallback
	if home, err := os.UserHomeDir(); err != nil {
		if !yield(``, err) {
			return
		}
	} else {
		if !yield(filepath.Join(home, `.local`, `share`), nil) {
			return
		}
	}
	if dataDir, ok := os.LookupEnv(`XDG_DATA_HOME`); ok {
		if !yield(dataDir, nil) {
			return
		}
	}
}
