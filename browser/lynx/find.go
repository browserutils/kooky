//go:build !windows
// +build !windows

// unix only

package lynx

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/cookies"
	"github.com/zellyn/kooky/internal/netscape"
)

type lynxFinder struct{}

var _ kooky.CookieStoreFinder = (*lynxFinder)(nil)

func init() {
	kooky.RegisterFinder(`lynx`, &lynxFinder{})
}

func (f *lynxFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	var ret []kooky.CookieStore

	// the default value is ~/.lynx_cookies for most systems, but ~/cookies for MS-DOS
	ret = append(
		ret,
		&cookies.CookieJar{
			CookieStore: &netscape.CookieStore{
				DefaultCookieStore: cookies.DefaultCookieStore{
					BrowserStr:           `lynx`,
					IsDefaultProfileBool: true,
					FileNameStr:          filepath.Join(home, `.lynx_cookies`),
				},
			},
		},
	)

	// parse config files so that we don't have to execute lynx -show_cfg
	configFiles := []string{
		filepath.Join(`/etc`, `lynx.cfg`),
		filepath.Join(`/etc`, `lynx`, `lynx.cfg`), // Debian
	}

	// INCLUDE:/etc/lynx/local.cfg
	// `/etc/lynx/lynx.cfg` includes `/etc/lynx/local.cfg` on Debian
	// https://lynx.invisible-island.net/current/README.cookies
	// COOKIE_FILE:/path/to/directory/.lynx_cookies // read file (?)
	// COOKIE_SAVE_FILE:/path/to/directory/.lynx_cookies // save file

	var includes, cookieFiles, cookieSaveFiles []string
	parse := func(configFile string) error {
		file, err := os.Open(configFile)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, `INCLUDE:`) {
				sp := strings.Split(line, `:`)
				if len(sp) == 2 {
					includes = append(includes, sp[1])
				}
			}
			if strings.HasPrefix(line, `COOKIE_FILE:`) {
				sp := strings.Split(line, `:`)
				if len(sp) == 2 {
					cookieFiles = append(cookieFiles, sp[1])
				}
			}
			if strings.HasPrefix(line, `COOKIE_SAVE_FILE:`) {
				sp := strings.Split(line, `:`)
				if len(sp) == 2 {
					cookieSaveFiles = append(cookieSaveFiles, sp[1])
				}
			}
		}
		return nil
	}

configFileLoop:
	for _, configFile := range configFiles {
		_ = parse(configFile)
	}
	if len(includes) > 0 {
		configFiles = includes
		includes = nil
		goto configFileLoop
	}

	var primCookieFile string
	if len(cookieFiles) > 0 {
		primCookieFile = cookieFiles[len(cookieFiles)-1]
	}

	cookieMap := make(map[string]struct{})
	for _, cookieFile := range append(cookieSaveFiles, cookieFiles...) {
		if _, exists := cookieMap[cookieFile]; exists {
			continue
		}
		cookieMap[cookieFile] = struct{}{}
	}

	{
		last := len(cookieMap) - 1
		i := 0
		for cookieFile := range cookieMap {
			ret = append(
				ret,
				&cookies.CookieJar{
					CookieStore: &netscape.CookieStore{
						DefaultCookieStore: cookies.DefaultCookieStore{
							BrowserStr: `lynx`,
							// last one probably overwrites earlier configuration
							IsDefaultProfileBool: cookieFile == primCookieFile || i == last,
							FileNameStr:          cookieFile,
						},
					},
				},
			)
			i++
		}
	}

	return ret, nil
}
