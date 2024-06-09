package netscape

import (
	"context"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/netscape"
)

// This ReadCookies() function returns an additional boolean "strict" telling
// if the file adheres to the netscape cookies.txt format
func ReadCookies(ctx context.Context, filename string, filters ...kooky.Filter) (_ []*kooky.Cookie, strict bool, _ error) {
	cs, str := TraverseCookies(filename, filters...)
	cookies, err := cs.ReadAllCookies(ctx)
	strict = err == nil && str()
	return cookies, strict, err
}

// This TraverseCookies() function returns an additional boolean returning func "strict" telling
// if the file adheres to the netscape cookies.txt format
func TraverseCookies(filename string, filters ...kooky.Filter) (_ kooky.CookieSeq, strict func() bool) {
	st := cookieStoreBasic(filename)
	stFilt := cookies.NewCookieJar(st, filters...)
	seq := cookies.ReadCookiesClose(stFilt, filters...)
	return seq, st.IsStrict
}

// CookieStore has to be closed with CookieStore.Close() after use.
func CookieStore(filename string, filters ...kooky.Filter) (kooky.CookieStore, error) {
	return cookieStore(filename, filters...)
}

func cookieStore(filename string, filters ...kooky.Filter) (*cookies.CookieJar, error) {
	return cookies.NewCookieJar(cookieStoreBasic(filename), filters...), nil
}

func cookieStoreBasic(filename string) *netscape.CookieStore {
	st := &netscape.CookieStore{}
	st.FileNameStr = filename
	st.BrowserStr = `netscape`
	return st
}

var ErrNotStrict = netscape.ErrNotStrict
