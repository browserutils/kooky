//go:build !windows
// +build !windows

package uzbl

import (
	"os"
	"path/filepath"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal"
	"github.com/zellyn/kooky/internal/netscape"
)

type uzblFinder struct{}

var _ kooky.CookieStoreFinder = (*uzblFinder)(nil)

func init() {
	kooky.RegisterFinder(`uzbl`, &uzblFinder{})
}

func (f *uzblFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	roots, err := uzblRoots()
	if err != nil {
		return nil, err
	}

	var ret []kooky.CookieStore

	for _, root := range roots {
		for _, filename := range []string{`session-cookies.txt`, `cookies.txt`} {
			var s netscape.CookieStore
			d := internal.DefaultCookieStore{
				BrowserStr:  `uzbl`,
				FileNameStr: filepath.Join(root, `uzbl`, filename),
			}
			internal.SetCookieStore(&d, &s)
			s.DefaultCookieStore = d
			ret = append(ret, &s)
		}
	}

	if len(ret) > 0 {
		if cookieStore, ok := ret[len(ret)-1].(*netscape.CookieStore); ok {
			cookieStore.IsDefaultProfileBool = true
		}
	}

	return ret, nil
}

func uzblRoots() ([]string, error) {
	var ret []string
	home, errHome := os.UserHomeDir()

	// old location
	// fallback
	if errHome == nil {
		ret = append(ret, filepath.Join(home, `.config`))
	}
	if dir, ok := os.LookupEnv(`XDG_CONFIG_HOME`); ok {
		ret = append(ret, dir)
	}

	// new location
	if errHome == nil {
		ret = append(ret, filepath.Join(home, `.local`, `share`))
	}

	if dataDir, ok := os.LookupEnv(`XDG_DATA_HOME`); ok {
		ret = append(ret, dataDir)
	}

	return ret, nil
}
