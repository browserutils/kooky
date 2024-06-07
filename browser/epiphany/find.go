//go:build !windows && !android && !ios

package epiphany

import (
	"os"
	"path/filepath"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
)

type epiphanyFinder struct{}

var _ kooky.CookieStoreFinder = (*epiphanyFinder)(nil)

func init() {
	kooky.RegisterFinder(`epiphany`, &epiphanyFinder{})
}

func (f *epiphanyFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		roots, err := epiphanyRoots()
		if err != nil {
			_ = yield(nil, err)
			return
		}

		last := len(roots) - 1
		for i, root := range roots {
			st := &cookies.CookieJar{
				CookieStore: &epiphanyCookieStore{
					DefaultCookieStore: cookies.DefaultCookieStore{
						BrowserStr:           `epiphany`,
						IsDefaultProfileBool: i == last,
						FileNameStr:          filepath.Join(root, `cookies.sqlite`),
					},
				},
			}
			if !yield(st, nil) {
				return
			}
		}

	}
}

func epiphanyRoots() ([]string, error) {
	var ret []string
	if home, err := os.UserHomeDir(); err == nil {
		ret = []string{
			filepath.Join(home, `.var`, `app`, `org.gnome.Epiphany`, `data`, `epiphany`), // flatpak
			filepath.Join(home, `.local`, `share`, `epiphany`),                           // fallback
		}
	}
	if dataDir, ok := os.LookupEnv(`XDG_DATA_HOME`); ok {
		ret = append(ret, filepath.Join(dataDir, `epiphany`))
	}

	return ret, nil
}
