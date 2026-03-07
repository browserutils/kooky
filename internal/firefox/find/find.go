package find

import (
	"errors"
	"iter"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

type firefoxCookieStoreFile struct {
	Path             string
	Browser          string
	Profile          string
	IsDefaultProfile bool
}

func FindFirefoxCookieStoreFiles() iter.Seq2[*firefoxCookieStoreFile, error] {
	return FindCookieStoreFiles(firefoxRoots, `firefox`, `cookies.sqlite`)
}

func FindCookieStoreFiles(rootsFunc iter.Seq2[string, error], browserName, fileName string) iter.Seq2[*firefoxCookieStoreFile, error] {
	return func(yield func(*firefoxCookieStoreFile, error) bool) {
		if rootsFunc == nil {
			_ = yield(nil, errors.New(`provided roots function is nil`))
			return
		}
		for root, err := range rootsFunc {
			if err != nil {
				if !yield(nil, err) {
					return
				}
				continue
			}
			iniFile := filepath.Join(root, `profiles.ini`)
			profIni, err := ini.Load(iniFile)
			if err != nil {
				continue
			}
			var defaultProfileFolder string
			var fallbackProfileFolder string
			var fallbackCount int
			for _, sec := range profIni.SectionStrings() {
				cfgSec := profIni.Section(sec)
				if cfgSec.Key(`Locked`).String() == `1` {
					defaultProfileFolder = cfgSec.Key(`Default`).String()
				}
				if strings.HasPrefix(sec, `Profile`) && cfgSec.Key(`Default`).String() == `1` {
					fallbackProfileFolder = cfgSec.Key(`Path`).String()
					fallbackCount++
				}
			}
			// fallback to Default=1 profile
			if defaultProfileFolder == `` && fallbackCount == 1 {
				defaultProfileFolder = fallbackProfileFolder
			}
			for _, sec := range profIni.SectionStrings() {
				// dedicated profiles (firefox 67+) start with Install instead of Profile followed by upper case hex
				// https://support.mozilla.org/en-US/kb/dedicated-profiles-firefox-installation
				if !strings.HasPrefix(sec, `Profile`) {
					continue
				}
				cfgSec := profIni.Section(sec)
				profileFolder := cfgSec.Key(`Path`).String()
				var defaultBrowser bool
				if defaultProfileFolder != `` && profileFolder == defaultProfileFolder {
					defaultBrowser = true
				}
				profileFolder = filepath.FromSlash(profileFolder)
				if cfgSec.Key(`IsRelative`).String() == `1` {
					// relative profile path
					profileFolder = filepath.Join(root, profileFolder)
				}
				st := &firefoxCookieStoreFile{
					Browser:          browserName,
					Profile:          cfgSec.Key(`Name`).String(),
					IsDefaultProfile: defaultBrowser,
					Path:             filepath.Join(profileFolder, fileName),
				}
				if !yield(st, nil) {
					return
				}
			}
		}
	}
}
