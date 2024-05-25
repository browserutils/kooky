package edge

import (
	"runtime"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/chrome"
	chromefind "github.com/browserutils/kooky/internal/chrome/find"
	"github.com/browserutils/kooky/internal/cookies"
)

type edgeFinder struct{}

var _ kooky.CookieStoreFinder = (*edgeFinder)(nil)

func init() {
	kooky.RegisterFinder(`edge`, &edgeFinder{})
}

func (f *edgeFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	files, err := chromefind.FindCookieStoreFiles(edgeChromiumRoots, `edge`)
	if err != nil {
		return nil, err
	}

	var ret []kooky.CookieStore
	for _, file := range files {
		cookieStore := &chrome.CookieStore{
			DefaultCookieStore: cookies.DefaultCookieStore{
				BrowserStr:           file.Browser,
				ProfileStr:           file.Profile,
				OSStr:                file.OS,
				IsDefaultProfileBool: file.IsDefaultProfile,
				FileNameStr:          file.Path,
			},
		}
		cookieStore.SetSafeStorage(`Microsoft Edge`, ``)
		ret = append(ret, &cookies.CookieJar{CookieStore: cookieStore})
	}

	var errRet error
	if runtime.GOOS != `windows` || edgeOldCookieStores == nil {
		goto skipNonChromium
	}
	{
		oldCookieStores, err := edgeOldCookieStores() // ESE, text cookies
		if err != nil {
			errRet = err
		}
		ret = append(ret, oldCookieStores...)
	}

skipNonChromium:
	return ret, errRet
}

var edgeOldCookieStores func() ([]kooky.CookieStore, error)
