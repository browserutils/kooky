//+build windows

package edge

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/chrome/find"
)

type edgeFinder struct{}

var _ kooky.CookieStoreFinder = (*edgeFinder)(nil)

func init() {
	kooky.RegisterFinder(`edge`, &edgeFinder{})
}

func (s *edgeFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	locApp := os.Getenv(`LocalAppData`)
	if len(locApp) == 0 {
		return nil, errors.New(`%LocalAppData% is empty`)
	}

	var cookiesFiles []kooky.CookieStore

	// old versions
	cookiesFiles = append(
		cookiesFiles,
		&edgeCookieStore{
			browser:  `edge`,
			filename: filepath.Join(locApp, `Microsoft`, `Windows`, `WebCache`, `WebCacheV01.dat`),
		},
	)

	// Blink based
	newRoot := func() ([]string, error) {
		return []string{filepath.Join(locApp, `Microsoft`, `Edge`, `User Data`)}, nil
	}
	if blinkCookiesFiles, err := find.FindCookieStoreFiles(newRoot, `edge`); err != nil {
		for _, cookiesFile := range blinkCookiesFiles {
			cookiesFiles = append(
				cookiesFiles,
				&edgeCookieStore{
					filename:         cookiesFile.Path,
					browser:          cookiesFile.Browser,
					profile:          cookiesFile.Profile,
					os:               cookiesFile.OS,
					isDefaultProfile: cookiesFile.IsDefaultProfile,
				},
			)
		}
	}

	return cookiesFiles, nil
}

/*
https://www.nirsoft.net/utils/edge_cookies_view.html
starting from Fall Creators Update 1709 of Windows 10, the cookies of Microsoft Edge Web browser are stored in the WebCacheV01.dat database
ESE database at %USERPROFILE%\AppData\Local\Microsoft\Windows\WebCache\WebCacheV01.dat (%LocalAppData%\Microsoft\Windows\WebCache\WebCacheV01.dat)
CookieEntryEx_##

https://www.linkedin.com/pulse/windows-10-microsoft-edge-browser-forensics-brent-muir
https://bsmuir.kinja.com/windows-10-microsoft-edge-browser-forensics-1733533818

locations:
%LocalAppData%\Microsoft\Windows\WebCache\WebCacheV01.dat
%LocalAppData%\Microsoft\Edge\User Data\Default

https://www.foxtonforensics.com/browser-history-examiner/microsoft-edge-history-location
v79+:
Edge Cookies are stored in the 'Cookies' SQLite database, within the 'cookies' table.

up to v44:
Edge Cookies are stored in the 'WebCacheV01.dat' ESE database, within the 'CookieEntryEx' containers.

older:
Older versions of Edge stored cookies as separate text files in locations specified within the ESE database.
*/
