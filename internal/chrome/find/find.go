package find

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
	"runtime"
)

type chromeCookieStoreFile struct {
	Path             string
	Browser          string
	Profile          string
	OS               string
	IsDefaultProfile bool
}

// chromeRoots and chromiumRoots could be put into the github.com/kooky/browser/{chrome,chromium} packages.
// It might be better though to keep those 2 together here as they are based on the same source.
func FindChromeCookieStoreFiles() ([]*chromeCookieStoreFile, error) {
	return FindCookieStoreFiles(chromeRoots, `chrome`)
}
func FindChromiumCookieStoreFiles() ([]*chromeCookieStoreFile, error) {
	return FindCookieStoreFiles(chromiumRoots, `chromium`)
}

func FindCookieStoreFiles(rootsFunc func() ([]string, error), browserName string) ([]*chromeCookieStoreFile, error) {
	if rootsFunc == nil {
		return nil, errors.New(`passed roots function is nil`)
	}
	var files []*chromeCookieStoreFile
	roots, err := rootsFunc()
	if err != nil {
		return nil, err
	}
	for _, root := range roots {
		localStateBytes, err := ioutil.ReadFile(filepath.Join(root, `Local State`))
		if err != nil {
			continue
		}
		var localState struct {
			Profile struct {
				InfoCache map[string]struct {
					IsUsingDefaultName bool `json:"is_using_default_name"`
					Name               string
				} `json:"info_cache"`
			}
		}
		if err := json.Unmarshal(localStateBytes, &localState); err != nil {
			// fallback - json file exists, json structure unknown
			files = append(
				files,
				[]*chromeCookieStoreFile{
					{
						Browser:          browserName,
						Profile:          `Profile 1`,
						IsDefaultProfile: true,
						Path:             filepath.Join(root, `Default`, `Network`, `Cookies`), // Chrome 96
						OS:               runtime.GOOS,
					},
					{
						Browser:          browserName,
						Profile:          `Profile 1`,
						IsDefaultProfile: true,
						Path:             filepath.Join(root, `Default`, `Cookies`),
						OS:               runtime.GOOS,
					},
				}...,
			)
			continue

		}
		for profDir, profStr := range localState.Profile.InfoCache {
			files = append(
				files,
				[]*chromeCookieStoreFile{
					{
						Browser:          browserName,
						Profile:          profStr.Name,
						IsDefaultProfile: profStr.IsUsingDefaultName,
						Path:             filepath.Join(root, profDir, `Network`, `Cookies`), // Chrome 96
						OS:               runtime.GOOS,
					}, {
						Browser:          browserName,
						Profile:          profStr.Name,
						IsDefaultProfile: profStr.IsUsingDefaultName,
						Path:             filepath.Join(root, profDir, `Cookies`),
						OS:               runtime.GOOS,
					},
				}...,
			)
		}
	}
	return files, nil
}
