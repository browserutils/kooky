package find

import (
	"errors"
	"iter"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

// Profile represents a Firefox-based browser profile discovered from profiles.ini.
type Profile struct {
	Path             string
	Browser          string
	Name             string
	IsDefaultProfile bool
}

// FindFirefoxProfiles returns all Firefox profiles from known root directories.
func FindFirefoxProfiles() iter.Seq2[Profile, error] {
	return FindProfiles(firefoxRoots, `firefox`)
}

// FindProfilesInRoot parses a single profiles.ini from rootDir
// and returns the discovered profiles.
func FindProfilesInRoot(rootDir, browserName string) ([]Profile, error) {
	iniFile := filepath.Join(rootDir, `profiles.ini`)
	profIni, err := ini.Load(iniFile)
	if err != nil {
		return nil, err
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
	var profiles []Profile
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
			profileFolder = filepath.Join(rootDir, profileFolder)
		}
		profiles = append(profiles, Profile{
			Browser:          browserName,
			Name:             cfgSec.Key(`Name`).String(),
			IsDefaultProfile: defaultBrowser,
			Path:             profileFolder,
		})
	}
	return profiles, nil
}

// FindProfiles lazily iterates root directories, parses profiles.ini in each,
// and yields discovered profiles.
func FindProfiles(rootsFunc iter.Seq2[string, error], browserName string) iter.Seq2[Profile, error] {
	return func(yield func(Profile, error) bool) {
		if rootsFunc == nil {
			_ = yield(Profile{}, errors.New(`provided roots function is nil`))
			return
		}
		for root, err := range rootsFunc {
			if err != nil {
				if !yield(Profile{}, err) {
					return
				}
				continue
			}
			profiles, err := FindProfilesInRoot(root, browserName)
			if err != nil {
				// profiles.ini not found or unparseable — skip this root
				continue
			}
			for _, p := range profiles {
				if !yield(p, nil) {
					return
				}
			}
		}
	}
}

// FindCookieStoreFiles lazily iterates profiles from the given root directories
// and yields cookie store file entries with fileName appended to each profile path.
// Used by netscape which has a different store type but reuses profile discovery.
func FindCookieStoreFiles(rootsFunc iter.Seq2[string, error], browserName, fileName string) iter.Seq2[*CookieStoreFile, error] {
	return func(yield func(*CookieStoreFile, error) bool) {
		for p, err := range FindProfiles(rootsFunc, browserName) {
			if err != nil {
				if !yield(nil, err) {
					return
				}
				continue
			}
			st := &CookieStoreFile{
				Browser:          p.Browser,
				Profile:          p.Name,
				IsDefaultProfile: p.IsDefaultProfile,
				Path:             filepath.Join(p.Path, fileName),
			}
			if !yield(st, nil) {
				return
			}
		}
	}
}

// CookieStoreFile represents a cookie store file discovered within a profile.
type CookieStoreFile struct {
	Path             string
	Browser          string
	Profile          string
	IsDefaultProfile bool
}
