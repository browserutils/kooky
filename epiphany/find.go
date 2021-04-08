//+build !windows

package epiphany

import (
	"os"
	"path/filepath"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal"
)

type epiphanyFinder struct{}

var _ kooky.CookieStoreFinder = (*epiphanyFinder)(nil)

func init() {
	kooky.RegisterFinder(`epiphany`, &epiphanyFinder{})
}

func (s *epiphanyFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	roots, err := epiphanyRoots()
	if err != nil {
		return nil, err
	}

	var ret []kooky.CookieStore

	for _, root := range roots {
		ret = append(
			ret,
			&epiphanyCookieStore{
				DefaultCookieStore: internal.DefaultCookieStore{
					BrowserStr:  `epiphany`,
					FileNameStr: filepath.Join(root, `epiphany`, `cookies.sqlite`),
				},
			},
		)
	}

	if len(ret) > 0 {
		if cookieStore, ok := ret[len(ret)-1].(*epiphanyCookieStore); ok {
			cookieStore.IsDefaultProfileBool = true
		}

	}

	return ret, nil
}

func epiphanyRoots() ([]string, error) {
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
