//go:build windows

package find

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/cookies"
	"github.com/zellyn/kooky/internal/ie"
)

type finder struct {
	browser string
}

var _ kooky.CookieStoreFinder = (*finder)(nil)

var registerOnce sync.Once

func init() {
	browser := `ie+edge`
	// don't register multiple times for files shared between ie and edge
	registerOnce.Do(func() {
		kooky.RegisterFinder(browser, &finder{browser: browser})
	})
}

func (f *finder) FindCookieStores() ([]kooky.CookieStore, error) {
	locApp := os.Getenv(`LOCALAPPDATA`)
	home := os.Getenv(`USERPROFILE`)
	windows := os.Getenv(`windir`)
	appData, _ := os.UserConfigDir()

	type pathStruct struct {
		dir   string
		paths [][]string
	}

	// https://tzworks.com/prototypes/index_dat/id.users.guide.pdf
	paths := []pathStruct{
		{
			dir: windows,
			paths: [][]string{
				[]string{`Cookies`}, // IE 4.0
			},
		},
		{
			dir: home,
			paths: [][]string{
				[]string{`Cookies`}, // XP, Vista
			},
		},
		{
			dir: appData,
			paths: [][]string{
				[]string{`Microsoft`, `Windows`, `Cookies`},
				[]string{`Microsoft`, `Windows`, `Cookies`, `Low`},
				[]string{`Microsoft`, `Windows`, `Cookies`, `Low`},
				[]string{`Microsoft`, `Windows`, `Internet Explorer`, `UserData`, `Low`},
			},
		},
	}

	var cookiesFiles []kooky.CookieStore
	for _, p := range paths {
		if len(p.dir) == 0 {
			continue
		}
		for _, path := range p.paths {
			cookiesFiles = append(
				cookiesFiles,
				&cookies.CookieJar{
					CookieStore: &ie.CookieStore{
						CookieStore: &ie.IECacheCookieStore{
							DefaultCookieStore: cookies.DefaultCookieStore{
								BrowserStr:           f.browser,
								IsDefaultProfileBool: true,
								FileNameStr:          filepath.Join(append(append([]string{p.dir}, path...), `index.dat`)...),
							},
						},
					},
				},
			)
		}
	}

	cookiesFiles = append(
		cookiesFiles,
		&cookies.CookieJar{
			CookieStore: &ie.CookieStore{
				CookieStore: &ie.ESECookieStore{
					DefaultCookieStore: cookies.DefaultCookieStore{
						BrowserStr:           f.browser,
						IsDefaultProfileBool: true,
						FileNameStr:          filepath.Join(locApp, `Microsoft`, `Windows`, `WebCache`, `WebCacheV01.dat`),
					},
				},
			},
		},
	)

	return cookiesFiles, nil
}
