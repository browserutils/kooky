//go:build !windows
// +build !windows

package uzbl

import (
	"iter"
	"os"
	"path/filepath"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/netscape"
)

type uzblFinder struct{}

var _ kooky.CookieStoreFinder = (*uzblFinder)(nil)

func init() {
	kooky.RegisterFinder(`uzbl`, &uzblFinder{})
}

// TODO default profile

func (f *uzblFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		files := []string{`session-cookies.txt`, `cookies.txt`}

		for root, err := range uzblRoots() {
			if err != nil && !yield(nil, err) {
				return
			}
			for _, filename := range files {
				st := &cookies.CookieJar{
					CookieStore: &netscape.CookieStore{
						DefaultCookieStore: cookies.DefaultCookieStore{
							BrowserStr:  `uzbl`,
							FileNameStr: filepath.Join(root, `uzbl`, filename),
						},
					},
				}
				if !yield(st, nil) {
					return
				}
			}
		}
	}
}

func uzblRoots() iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		home, errHome := os.UserHomeDir()

		// old location
		// fallback
		if errHome != nil {
			if !yield(``, errHome) {
				return
			}
		} else {
			if !yield(filepath.Join(home, `.config`), nil) {
				return
			}
		}
		if dir, ok := os.LookupEnv(`XDG_CONFIG_HOME`); ok {
			if !yield(dir, nil) {
				return
			}
		}

		// new location
		if errHome == nil {
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
}
