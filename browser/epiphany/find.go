//go:build !windows && !android && !ios

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
					FileNameStr: filepath.Join(root, `cookies.sqlite`),
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
