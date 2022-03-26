package chrome

import (
	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal"
	"github.com/zellyn/kooky/internal/chrome"
	"github.com/zellyn/kooky/internal/chrome/find"
)

type chromiumFinder struct{}

var _ kooky.CookieStoreFinder = (*chromiumFinder)(nil)

func init() {
	kooky.RegisterFinder(`chromium`, &chromiumFinder{})
}

func (f *chromiumFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	files, err := find.FindChromiumCookieStoreFiles()
	if err != nil {
		return nil, err
	}

	var ret []kooky.CookieStore
	for _, file := range files {
		var s chrome.CookieStore
		d := internal.DefaultCookieStore{
			BrowserStr:           file.Browser,
			ProfileStr:           file.Profile,
			OSStr:                file.OS,
			IsDefaultProfileBool: file.IsDefaultProfile,
			FileNameStr:          file.Path,
		}
		internal.SetCookieStore(&d, &s)
		s.DefaultCookieStore = d
		ret = append(ret, &s)
	}

	return ret, nil
}
