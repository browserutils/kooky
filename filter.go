package kooky

import (
	"context"
	"errors"
	"net/http"
	"reflect"
)

// Filter is used for filtering cookies in ReadCookies() functions.
// Filter order might be changed for performance reasons
// (omission of value decryption of filtered out cookies, etc).
//
// A cookie passes the Filter if Filter.Filter returns true.
type Filter interface{ Filter(*Cookie) bool }

func FilterCookies[S CookieSeq | ~[]*Cookie | ~[]*http.Cookie](ctx context.Context, cookies S, filters ...Filter) CookieSeq {
	var ret CookieSeq
	// https://github.com/golang/go/issues/45380#issuecomment-1014950980
	switch cookiesTyped := any(cookies).(type) {
	case CookieSeq:
		ret = filterCookieSeq(ctx, cookiesTyped, filters...)
	case Cookies:
		ret = filterCookieSlice(ctx, []*Cookie(cookiesTyped), filters...)
	case []*Cookie:
		ret = filterCookieSlice(ctx, cookiesTyped, filters...)
	case []*http.Cookie:
		ret = filterCookieSlice(ctx, cookiesTyped, filters...)
	default:
		rv := reflect.ValueOf(cookies)
		rtc := reflect.TypeFor[[]*Cookie]()
		rthc := reflect.TypeFor[[]*http.Cookie]()
		if rv.CanConvert(rtc) {
			cookiesTyped := rv.Convert(rtc).Interface().([]*Cookie)
			ret = filterCookieSlice(ctx, cookiesTyped, filters...)
		} else if rv.CanConvert(rthc) {
			cookiesTyped := rv.Convert(rthc).Interface().([]*http.Cookie)
			ret = filterCookieSlice(ctx, cookiesTyped, filters...)
		} else {
			ret = func(yield func(*Cookie, error) bool) { yield(nil, errors.New(`unknown type`)) }
		}
	}
	return ret
}

func matchCookie(ctx context.Context, cookie *Cookie, filters ...Filter) bool {
	if cookie == nil {
		return false
	}
	for _, filter := range filters {
		if filter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			return false
		default:
		}
		if !filter.Filter(cookie) {
			return false
		}
	}
	return true
}

func filterCookieSeq(ctx context.Context, cookies CookieSeq, filters ...Filter) CookieSeq {
	return func(yield func(*Cookie, error) bool) {
		if cookies == nil {
			yield(nil, errors.New(`nil receiver`))
			return
		}
		for cookie, errCookie := range cookies {
			if errCookie != nil {
				if !yield(nil, errCookie) {
					return
				}
				continue
			}
			if cookie == nil {
				continue
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
			if !matchCookie(ctx, cookie, filters...) {
				continue
			}
			if !yield(cookie, nil) {
				return
			}
		}
	}
}

func filterCookieSlice[S ~[]*T, T Cookie | http.Cookie](ctx context.Context, cookies S, filters ...Filter) CookieSeq {
	return func(yield func(*Cookie, error) bool) {
		if len(cookies) < 1 {
			_ = yield(nil, errors.New(`cookie slice of lenght 0`))
			return
		}
		switch cookiesTyped := any(cookies).(type) {
		case []*http.Cookie:
		cookieLoopHTTP:
			for _, cookie := range cookiesTyped {
				if cookie == nil {
					continue
				}
				select {
				case <-ctx.Done():
					break cookieLoopHTTP
				default:
				}
				kooky := &Cookie{Cookie: *cookie}
				if !matchCookie(ctx, kooky, filters...) {
					continue
				}
				if !yield(kooky, nil) {
					return
				}
			}
		case []*Cookie:
		cookieLoopKooky:
			for _, cookie := range cookiesTyped {
				if cookie == nil {
					continue
				}
				select {
				case <-ctx.Done():
					break cookieLoopKooky
				default:
				}
				if !matchCookie(ctx, cookie, filters...) {
					continue
				}
				if !yield(cookie, nil) {
					return
				}
			}
		}
	}
}
