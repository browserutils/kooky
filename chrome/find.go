package chrome

import (
	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal"
	"github.com/zellyn/kooky/internal/chrome"
	"github.com/zellyn/kooky/internal/chrome/find"
)

type chromeFinder struct{}
type chromiumFinder struct{}

var _ kooky.CookieStoreFinder = (*chromeFinder)(nil)
var _ kooky.CookieStoreFinder = (*chromiumFinder)(nil)

func init() {
	kooky.RegisterFinder(`chrome`, &chromeFinder{})
	kooky.RegisterFinder(`chromium`, &chromiumFinder{})
}

func (s *chromeFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	files, err := find.FindChromeCookieStoreFiles()
	if err != nil {
		return nil, err
	}

	var ret []kooky.CookieStore
	for _, file := range files {
		ret = append(
			ret,
			&chrome.CookieStore{
				DefaultCookieStore: internal.DefaultCookieStore{
					BrowserStr:           file.Browser,
					ProfileStr:           file.Profile,
					OSStr:                file.OS,
					IsDefaultProfileBool: file.IsDefaultProfile,
					FileNameStr:          file.Path,
				},
			},
		)
	}

	return ret, nil
}

func (s *chromiumFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	files, err := find.FindChromiumCookieStoreFiles()
	if err != nil {
		return nil, err
	}

	var ret []kooky.CookieStore
	for _, file := range files {
		ret = append(
			ret,
			&chrome.CookieStore{
				DefaultCookieStore: internal.DefaultCookieStore{
					BrowserStr:           file.Browser,
					ProfileStr:           file.Profile,
					OSStr:                file.OS,
					IsDefaultProfileBool: file.IsDefaultProfile,
					FileNameStr:          file.Path,
				},
			},
		)
	}

	return ret, nil
}
