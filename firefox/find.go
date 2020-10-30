package firefox

import (
	"github.com/zellyn/kooky"
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
			&firefoxCookieStore{
				filename:         file.Path,
				browser:          file.Browser,
				profile:          file.Profile,
				isDefaultProfile: file.IsDefaultProfile,
			},
		)
	}

	return ret, nil
}
