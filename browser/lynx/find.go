//go:build !windows
// +build !windows

// unix only

package lynx

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/netscape"
	"github.com/browserutils/kooky/internal/utils"
)

type lynxFinder struct{}

var _ kooky.CookieStoreFinder = (*lynxFinder)(nil)

func init() {
	kooky.RegisterFinder(`lynx`, &lynxFinder{})
}

func (f *lynxFinder) FindCookieStores() kooky.CookieStoreSeq {
	return func(yield func(kooky.CookieStore, error) bool) {
		home, err := os.UserHomeDir()
		if err != nil {
			_ = yield(nil, err)
			return
		}

		// the default value is ~/.lynx_cookies for most systems, but ~/cookies for MS-DOS
		st := &cookies.CookieJar{
			CookieStore: &netscape.CookieStore{
				DefaultCookieStore: cookies.DefaultCookieStore{
					BrowserStr:           `lynx`,
					IsDefaultProfileBool: true,
					FileNameStr:          filepath.Join(home, `.lynx_cookies`),
				},
			},
		}
		if !yield(st, nil) {
			return
		}

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

		var primCookieFile string
		storeForFile := func(cookieFile string) *cookies.CookieJar {
			return &cookies.CookieJar{
				CookieStore: &netscape.CookieStore{
					DefaultCookieStore: cookies.DefaultCookieStore{
						BrowserStr: `lynx`,
						// last one probably overwrites earlier configuration
						IsDefaultProfileBool: cookieFile == primCookieFile,
						FileNameStr:          cookieFile,
					},
				},
			}
		}

		cookieMap := make(map[string]struct{})
		var includes, cookieFiles, cookieSaveFiles []string
		parse := func(configFile string) error {
			file, err := utils.OpenFile(configFile)
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
						// cookie file
						cookieFile := sp[1]
						primCookieFile = cookieFile
						cookieFiles = append(cookieFiles, cookieFile)
						cookieMap[cookieFile] = struct{}{}
					}
				}
				if strings.HasPrefix(line, `COOKIE_SAVE_FILE:`) {
					sp := strings.Split(line, `:`)
					if len(sp) == 2 {
						// cookie save file
						cookieSaveFile := sp[1]
						cookieSaveFiles = append(cookieSaveFiles, cookieSaveFile)
						cookieMap[cookieSaveFile] = struct{}{}
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

		// primCookieFile is now set
		for cookieFile := range cookieMap {
			if !yield(storeForFile(cookieFile), nil) {
				return
			}
		}
	}
}
