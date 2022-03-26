package firefox

import (
	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal"
	"github.com/zellyn/kooky/internal/firefox"
	"github.com/zellyn/kooky/internal/firefox/find"
)

type firefoxFinder struct{}

var _ kooky.CookieStoreFinder = (*firefoxFinder)(nil)

func init() {
	kooky.RegisterFinder(`firefox`, &firefoxFinder{})
}

func (f *firefoxFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	files, err := find.FindFirefoxCookieStoreFiles()
	if err != nil {
		return nil, err
	}

	var ret []kooky.CookieStore
	for _, file := range files {
		var s firefox.CookieStore
		d := internal.DefaultCookieStore{
			BrowserStr:           file.Browser,
			ProfileStr:           file.Profile,
			IsDefaultProfileBool: file.IsDefaultProfile,
			FileNameStr:          file.Path,
		}
		internal.SetCookieStore(&d, &s)
		s.DefaultCookieStore = d
		ret = append(ret, &s)
	}

	return ret, nil
}
