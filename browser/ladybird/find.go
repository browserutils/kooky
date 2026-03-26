//go:build !windows && !android && !ios

package ladybird

import (
	"os"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
)

type ladybirdFinder struct{}

var _ kooky.CookieStoreFinder = (*ladybirdFinder)(nil)

func init() {
	kooky.RegisterFinder(`ladybird`, &ladybirdFinder{})
}

func (f *ladybirdFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		for _, dbPath := range ladybirdCookiePaths() {
			if _, err := os.Stat(dbPath); err != nil {
				continue
			}
			st := &cookies.CookieJar{
				CookieStore: &ladybirdCookieStore{
					DefaultCookieStore: cookies.DefaultCookieStore{
						BrowserStr:           `ladybird`,
						IsDefaultProfileBool: true,
						FileNameStr:          dbPath,
					},
				},
			}
			if !yield(st, nil) {
				return
			}
		}
	}
}
