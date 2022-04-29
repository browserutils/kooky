//go:build !windows
// +build !windows

package uzbl

import (
	"os"
	"path/filepath"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/cookies"
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
	files := []string{`session-cookies.txt`, `cookies.txt`}

	var ret []kooky.CookieStore
	lastRoot := len(roots) - 1
	lastFile := len(files) - 1
	for i, root := range roots {
		for j, filename := range files {
			ret = append(
				ret,
				&cookies.CookieJar{
					CookieStore: &netscape.CookieStore{
						DefaultCookieStore: cookies.DefaultCookieStore{
							BrowserStr:           `uzbl`,
							IsDefaultProfileBool: i == lastRoot && j == lastFile,
							FileNameStr:          filepath.Join(root, `uzbl`, filename),
						},
					},
				},
			)
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
