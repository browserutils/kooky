package opera

import (
	"path/filepath"

	"github.com/zellyn/kooky"
)

type operaFinder struct{}

var _ kooky.CookieStoreFinder = (*operaFinder)(nil)

func init() {
	kooky.RegisterFinder(`opera`, &operaFinder{})
}

func (s *operaFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	var files []kooky.CookieStore

	roots, err := operaPrestoRoots()
	if err != nil {
		return nil, err
	}
	for _, root := range roots {
		files = append(
			files,
			&operaCookieStore{
				browser:          `opera`,
				isDefaultProfile: true,
				filename:         filepath.Join(root, `cookies4.dat`),
			},
		)
	}

	roots, err = operaBlinkRoots()
	if err != nil {
		return nil, err
	}
	for _, root := range roots {
		files = append(
			files,
			&operaCookieStore{
				browser:          `opera`,
				isDefaultProfile: true,
				filename:         filepath.Join(root, `Cookies`),
			},
		)
	}

	return files, nil
}
