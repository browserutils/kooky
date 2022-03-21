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

func (s *firefoxFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	files, err := find.FindFirefoxCookieStoreFiles()
	if err != nil {
		return nil, err
	}

	var ret []kooky.CookieStore
	for _, file := range files {
		ret = append(
			ret,
			&firefox.CookieStore{
				DefaultCookieStore: internal.DefaultCookieStore{
					BrowserStr:           file.Browser,
					ProfileStr:           file.Profile,
					IsDefaultProfileBool: file.IsDefaultProfile,
					FileNameStr:          file.Path,
				},
			},
		)
	}

	return ret, nil
}
