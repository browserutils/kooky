package edge

import (
	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/chrome"
	"github.com/browserutils/kooky/internal/chrome/find"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/edge"
	edgefind "github.com/browserutils/kooky/internal/edge/find"
)

// TODO !windows platforms

type edgeFinder struct{}

var _ kooky.CookieStoreFinder = (*edgeFinder)(nil)

func init() {
	kooky.RegisterFinder(`edge`, &edgeFinder{})
}

func (f *edgeFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	files, err := find.FindCookieStoreFiles(edgefind.GetEdgeRoots(), `edge`)
	if err != nil {
		return nil, err
	}

	var ret []kooky.CookieStore
	for _, file := range files {
		ret = append(
			ret,
			&cookies.CookieJar{
				CookieStore: &edge.CookieStore{
					CookieStore: chrome.CookieStore{
						DefaultCookieStore: cookies.DefaultCookieStore{
							BrowserStr:           file.Browser,
							ProfileStr:           file.Profile,
							OSStr:                file.OS,
							IsDefaultProfileBool: file.IsDefaultProfile,
							FileNameStr:          file.Path,
						},
					},
				},
			},
		)
	}
	return ret, nil
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
