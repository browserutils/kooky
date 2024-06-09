//go:build windows
// +build windows

package edge

import (
	"os"
	"path/filepath"

	iefind "github.com/browserutils/kooky/internal/ie/find"
)

func edgeChromiumRoots(yield func(string, error) bool) {
	// %LocalAppData%
	locApp, err := os.UserCacheDir()
	if err != nil {
		_ = yield(``, err)
		return
	}
	if !yield(filepath.Join(locApp, `Microsoft`, `Edge`, `User Data`), nil) {
		return
	}
}

func init() {
	edgeOldCookieStores = (&iefind.IEFinder{Browser: `edge`}).FindCookieStores()
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
