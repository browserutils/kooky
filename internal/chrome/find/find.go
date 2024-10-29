package find

import (
	"encoding/json"
	"errors"
	"iter"
	"os"
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
func FindChromeCookieStoreFiles() iter.Seq2[*chromeCookieStoreFile, error] {
	return FindCookieStoreFiles(chromeRoots, `chrome`)
}
func FindChromiumCookieStoreFiles() iter.Seq2[*chromeCookieStoreFile, error] {
	return FindCookieStoreFiles(chromiumRoots, `chromium`)
}

func FindCookieStoreFiles(rootsFunc iter.Seq2[string, error], browserName string) iter.Seq2[*chromeCookieStoreFile, error] {
	return func(yield func(*chromeCookieStoreFile, error) bool) {
		if rootsFunc == nil {
			_ = yield(nil, errors.New(`passed roots function is nil`))
			return
		}
		for root, err := range rootsFunc {
			if err != nil {
				if !yield(nil, err) {
					return
				}
				continue
			}
			localStateBytes, err := os.ReadFile(filepath.Join(root, `Local State`))
			if err != nil {
				if !yield(nil, err) {
					return
				}
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
				if !yield(nil, err) {
					return
				}
				st := &chromeCookieStoreFile{
					Browser:          browserName,
					Profile:          `Profile 1`,
					IsDefaultProfile: true,
					Path:             filepath.Join(root, `Default`, `Network`, `Cookies`), // Chrome 96
					OS:               runtime.GOOS,
				}
				if !yield(st, nil) {
					return
				}
				st = &chromeCookieStoreFile{
					Browser:          browserName,
					Profile:          `Profile 1`,
					IsDefaultProfile: true,
					Path:             filepath.Join(root, `Default`, `Cookies`),
					OS:               runtime.GOOS,
				}
				if !yield(st, nil) {
					return
				}
				continue
			}
			for profDir, profStr := range localState.Profile.InfoCache {
				st := &chromeCookieStoreFile{

					Browser:          browserName,
					Profile:          profStr.Name,
					IsDefaultProfile: profStr.IsUsingDefaultName,
					Path:             filepath.Join(root, profDir, `Network`, `Cookies`), // Chrome 96
					OS:               runtime.GOOS,
				}
				if !yield(st, nil) {
					return
				}
				st = &chromeCookieStoreFile{
					Browser:          browserName,
					Profile:          profStr.Name,
					IsDefaultProfile: profStr.IsUsingDefaultName,
					Path:             filepath.Join(root, profDir, `Cookies`),
					OS:               runtime.GOOS,
				}
				if !yield(st, nil) {
					return
				}
			}
		}
	}
}
