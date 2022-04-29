package ie

import (
	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/cookies"
	"github.com/zellyn/kooky/internal/ie"
	_ "github.com/zellyn/kooky/internal/ie/find"
)

type ieFinder struct{}

var _ kooky.CookieStoreFinder = (*ieFinder)(nil)

func init() {
	kooky.RegisterFinder(`ie`, &ieFinder{})
}

func (f *ieFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	roots, _ := ieRoots()
	var cookiesFiles []kooky.CookieStore
	for _, root := range roots {
		cookiesFiles = append(
			cookiesFiles,
			&cookies.CookieJar{
				CookieStore: &ie.CookieStore{
					CookieStore: &ie.TextCookieStore{
						DefaultCookieStore: cookies.DefaultCookieStore{
							BrowserStr:           `ie`,
							IsDefaultProfileBool: true,
							FileNameStr:          root,
						},
					},
				},
			},
		)
	}

	return cookiesFiles, nil
}
