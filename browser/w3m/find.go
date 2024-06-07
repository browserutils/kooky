//go:build linux || freebsd || openbsd || netbsd || dragonfly || solaris

package w3m

import (
	"os"
	"path/filepath"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
)

type w3mFinder struct{}

var _ kooky.CookieStoreFinder = (*w3mFinder)(nil)

func init() {
	kooky.RegisterFinder(`w3m`, &w3mFinder{})
}

func (f *w3mFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		home, err := os.UserHomeDir()
		if err != nil {
			_ = yield(nil, err)
			return
		}

		st := &cookies.CookieJar{
			CookieStore: &w3mCookieStore{
				DefaultCookieStore: cookies.DefaultCookieStore{
					BrowserStr:           `w3m`,
					IsDefaultProfileBool: true,
					FileNameStr:          filepath.Join(home, `.w3m`, `cookie`),
				},
			},
		}
		if !yield(st, nil) {
			return
		}
	}
}
