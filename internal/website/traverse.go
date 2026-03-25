//go:build js

package website

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"strings"
	"syscall/js"
	"time"

	"github.com/browserutils/kooky"
)

// TraverseCookies yields cookies using the Cookie Store API
// with a fallback to document.cookie. If apiUsed is non-nil, it is
// set to the name of the API that provided the cookies.
func TraverseCookies(browser kooky.BrowserInfo, apiUsed *string, filters ...kooky.Filter) iter.Seq2[Cookie, error] {
	return func(yield func(Cookie, error) bool) {
		if cookieStoreAPI(browser, yield, filters...) {
			if apiUsed != nil {
				*apiUsed = "cookieStore"
			}
			return
		}
		if apiUsed != nil {
			*apiUsed = "document.cookie"
		}
		documentCookie(browser, yield, filters...)
	}
}

// cookieStoreAPI attempts to use the Cookie Store API (cookieStore.getAll()).
// Returns true if the API was available and succeeded.
// TODO: extract exact Name filter and pass as getAll({name: "..."})
// and compose Domain+Path filters into getAll({url: "..."}) to narrow the JS-side query
// (scheme can be derived from location.protocol).
func cookieStoreAPI(browser kooky.BrowserInfo, yield func(Cookie, error) bool, filters ...kooky.Filter) bool {
	global := js.Global()
	store := global.Get("cookieStore")
	if store.IsUndefined() || store.IsNull() {
		return false
	}

	getAllFunc := store.Get("getAll")
	if getAllFunc.IsUndefined() || getAllFunc.IsNull() {
		return false
	}

	promise := store.Call("getAll")
	result, err := awaitPromise(promise)
	if err != nil {
		yield(Cookie{}, fmt.Errorf("cookieStore.getAll: %w", err))
		return false
	}

	// location values used as fallback when the API returns null fields.
	// some browsers (e.g. Firefox) return null for domain/path on host-only cookies.
	locDomain, locPath, _ := locationInfo(global)

	length := result.Length()
	for i := range length {
		entry := result.Index(i)
		var d DerivedFields

		cookie := &kooky.Cookie{}
		cookie.Name = jsString(entry, "name")
		cookie.Value = jsString(entry, "value")
		cookie.Domain = jsString(entry, "domain")
		cookie.Path = jsString(entry, "path")
		cookie.Secure = jsBool(entry, "secure")
		cookie.Partitioned = jsBool(entry, "partitioned")
		cookie.Browser = browser

		// fill missing domain from location.hostname.
		// host-only cookies (set without Domain= attribute) return null from the API.
		if cookie.Domain == "" {
			cookie.Domain = locDomain
			d.Domain = true
		}
		// fill missing path from location.pathname.
		// this is the current page path — the actual cookie path may be a parent path.
		if cookie.Path == "" {
			cookie.Path = locPath
			d.Path = true
		}
		// do NOT infer Secure from protocol — the Cookie Store API
		// returns the actual Secure attribute as set by the server.

		if expiresVal := entry.Get("expires"); !expiresVal.IsNull() && !expiresVal.IsUndefined() {
			// expires is Unix time in milliseconds
			cookie.Expires = time.UnixMilli(int64(expiresVal.Float()))
		}

		cookie.SameSite = ParseSameSite(jsString(entry, "sameSite"))

		if kooky.FilterCookie(context.Background(), cookie, filters...) {
			if !yield(Cookie{Cookie: cookie, Derived: d}, nil) {
				return true
			}
		}
	}
	return true
}

// locationInfo returns domain, path, and secure inferred from window.location.
func locationInfo(global js.Value) (domain, path string, secure bool) {
	location := global.Get("location")
	if !location.IsUndefined() && !location.IsNull() {
		domain = jsString(location, "hostname")
		path = jsString(location, "pathname")
		secure = jsString(location, "protocol") == "https:"
	}
	return
}

// documentCookie parses the document.cookie string.
// Only name and value are available directly;
// domain, path, and secure are all inferred from window.location.
func documentCookie(browser kooky.BrowserInfo, yield func(Cookie, error) bool, filters ...kooky.Filter) {
	global := js.Global()
	document := global.Get("document")
	if document.IsUndefined() || document.IsNull() {
		yield(Cookie{}, errors.New(`document is not available`))
		return
	}

	cookieVal := document.Get("cookie")
	if cookieVal.IsUndefined() || cookieVal.IsNull() {
		return
	}
	cookieStr := cookieVal.String()
	if len(cookieStr) == 0 {
		return
	}

	domain, path, secure := locationInfo(global)

	for pair := range splitCookieStr(cookieStr) {
		name, value, _ := strings.Cut(pair, "=")
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		if len(name) == 0 {
			continue
		}

		cookie := &kooky.Cookie{}
		cookie.Name = name
		cookie.Value = value
		cookie.Domain = domain
		cookie.Path = path
		cookie.Secure = secure
		cookie.Browser = browser

		// document.cookie provides only name=value;
		// all other fields are inferred from location.
		d := DerivedFields{Domain: true, Path: true, Secure: true}

		if kooky.FilterCookie(context.Background(), cookie, filters...) {
			if !yield(Cookie{Cookie: cookie, Derived: d}, nil) {
				return
			}
		}
	}
}

// splitCookieStr yields trimmed, non-empty cookie pair strings
// from a semicolon-separated document.cookie value.
func splitCookieStr(s string) func(yield func(string) bool) {
	return func(yield func(string) bool) {
		for {
			var part string
			var found bool
			part, s, found = strings.Cut(s, ";")
			part = strings.TrimSpace(part)
			if len(part) > 0 {
				if !yield(part) {
					return
				}
			}
			if !found {
				return
			}
		}
	}
}

func awaitPromise(promise js.Value) (js.Value, error) {
	type result struct {
		val js.Value
		err error
	}
	ch := make(chan result, 1)

	thenFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		var val js.Value
		if len(args) > 0 {
			val = args[0]
		}
		ch <- result{val: val}
		return nil
	})
	defer thenFunc.Release()

	catchFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		var errMsg string
		if len(args) > 0 {
			errMsg = args[0].Call("toString").String()
		} else {
			errMsg = "unknown promise rejection"
		}
		ch <- result{err: errors.New(errMsg)}
		return nil
	})
	defer catchFunc.Release()

	promise.Call("then", thenFunc).Call("catch", catchFunc)

	r := <-ch
	return r.val, r.err
}

func jsString(v js.Value, prop string) string {
	p := v.Get(prop)
	if p.IsUndefined() || p.IsNull() {
		return ``
	}
	return p.String()
}

func jsBool(v js.Value, prop string) bool {
	p := v.Get(prop)
	if p.IsUndefined() || p.IsNull() {
		return false
	}
	return p.Bool()
}
