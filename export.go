package kooky

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	httpOnlyPrefix = `#HttpOnly_`
	netscapeHeader = "# HTTP Cookie File\n\n"
)

// ExportCookies() export "cookies" in the Netscape format.
//
// curl, wget, ... use this format.
func ExportCookies[S CookieSeq | []*Cookie | []*http.Cookie](ctx context.Context, w io.Writer, cookies S) {
	switch cookiesTyped := any(cookies).(type) {
	case CookieSeq:
		exportCookieSeq(ctx, w, cookiesTyped)
	case []*Cookie:
		exportCookieSlice(ctx, w, cookiesTyped)
	case []*http.Cookie:
		exportCookieSlice(ctx, w, cookiesTyped)
	}
}

func exportCookie(w io.Writer, cookie *http.Cookie) {
	var domain string
	if cookie.HttpOnly {
		domain = httpOnlyPrefix
	}
	domain += cookie.Domain

	fmt.Fprintf(
		w,
		"%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
		domain,
		netscapeBool(strings.HasPrefix(cookie.Domain, `.`)),
		cookie.Path,
		netscapeBool(cookie.Secure),
		cookie.Expires.Unix(),
		cookie.Name,
		cookie.Value,
	)
}

func exportCookieSlice[S []*T, T Cookie | http.Cookie](ctx context.Context, w io.Writer, cookies S) {
	if len(cookies) < 1 {
		return
	}
	var j int
	for i, cookie := range cookies {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if cookie == nil {
			continue
		}
		fmt.Fprint(w, netscapeHeader)
		j = i
		break
	}

	// https://github.com/golang/go/issues/45380#issuecomment-1014950980
	switch cookiesTyp := any(cookies).(type) {
	case CookieSeq:
	case []*http.Cookie:
		for i := j; i < len(cookiesTyp); i++ {
			if cookiesTyp[i] == nil {
				continue
			}
			exportCookie(w, cookiesTyp[i])
		}
	case []*Cookie:
		for i := j; i < len(cookiesTyp); i++ {
			if cookiesTyp[i] == nil {
				continue
			}
			exportCookie(w, &cookiesTyp[i].Cookie)
		}
	}
}

func exportCookieSeq(ctx context.Context, w io.Writer, seq CookieSeq) {
	var init bool
	for cookie := range seq.OnlyCookies() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if !init {
			fmt.Fprint(w, netscapeHeader)
			init = true
		}
		exportCookie(w, &cookie.Cookie)
	}
}

func (s CookieSeq) Export(ctx context.Context, w io.Writer) { exportCookieSeq(ctx, w, s) }

type netscapeBool bool

func (b netscapeBool) String() string {
	if b {
		return `TRUE`
	}
	return `FALSE`
}
