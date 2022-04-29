//go:build !windows

package konqueror

import (
	"os"
	"path/filepath"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/cookies"
)

type konquerorFinder struct{}

var _ kooky.CookieStoreFinder = (*konquerorFinder)(nil)

func init() {
	kooky.RegisterFinder(`konqueror`, &konquerorFinder{})
}

func (f *konquerorFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	roots, err := konquerorRoots()
	if err != nil {
		return nil, err
	}

	var ret []kooky.CookieStore

	last := len(roots) - 1
	for i, root := range roots {
		ret = append(
			ret,
			&cookies.CookieJar{
				CookieStore: &konquerorCookieStore{
					DefaultCookieStore: cookies.DefaultCookieStore{
						BrowserStr:           `konqueror`,
						IsDefaultProfileBool: i == last,
						FileNameStr:          filepath.Join(root, `kcookiejar`, `cookies`),
					},
				},
			},
		)
	}

	return ret, nil
}

func konquerorRoots() ([]string, error) {
	var ret []string
	// fallback
	if home, err := os.UserHomeDir(); err == nil {
		ret = append(ret, filepath.Join(home, `.local`, `share`))
	}
	if dataDir, ok := os.LookupEnv(`XDG_DATA_HOME`); ok {
		ret = append(ret, dataDir)
	}

	return ret, nil
}
