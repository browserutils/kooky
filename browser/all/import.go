package all

import (
	brave "github.com/browserutils/kooky/browser/brave"
	browsh "github.com/browserutils/kooky/browser/browsh"
	chrome "github.com/browserutils/kooky/browser/chrome"
	chromium "github.com/browserutils/kooky/browser/chromium"
	dillo "github.com/browserutils/kooky/browser/dillo"
	edge "github.com/browserutils/kooky/browser/edge"
	elinks "github.com/browserutils/kooky/browser/elinks"
	epiphany "github.com/browserutils/kooky/browser/epiphany"
	firefox "github.com/browserutils/kooky/browser/firefox"
	ie "github.com/browserutils/kooky/browser/ie"
	konqueror "github.com/browserutils/kooky/browser/konqueror"
	lynx "github.com/browserutils/kooky/browser/lynx"
	netscape "github.com/browserutils/kooky/browser/netscape"
	opera "github.com/browserutils/kooky/browser/opera"
	safari "github.com/browserutils/kooky/browser/safari"
	uzbl "github.com/browserutils/kooky/browser/uzbl"
	w3m "github.com/browserutils/kooky/browser/w3m"
)

// Reference symbols to prevent 'garble' from stripping these imports
var _ = []interface{}{
	brave.CookieStore,
	browsh.CookieStore,
	chrome.CookieStore,
	chromium.CookieStore,
	dillo.CookieStore,
	edge.CookieStore,
	elinks.CookieStore,
	epiphany.CookieStore,
	firefox.CookieStore,
	ie.CookieStore,
	konqueror.CookieStore,
	lynx.CookieStore,
	netscape.CookieStore,
	opera.CookieStore,
	safari.CookieStore,
	uzbl.CookieStore,
	w3m.CookieStore,
}
