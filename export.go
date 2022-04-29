package kooky

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

const httpOnlyPrefix = `#HttpOnly_`

// ExportCookies() export "cookies" in the Netscape format.
//
// curl, wget, ... use this format.
func ExportCookies[T Cookie | http.Cookie](w io.Writer, cookies []*T) {
	if len(cookies) < 1 {
		return
	}
	var j int
	for i, cookie := range cookies {
		if cookie == nil {
			continue
		}
		fmt.Fprint(w, "# HTTP Cookie File\n\n")
		j = i
		break
	}

	writeCookie := func(w io.Writer, cookie *http.Cookie) {
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

	// https://github.com/golang/go/issues/45380#issuecomment-1014950980
	switch cookiesTyp := any(cookies).(type) {
	case []*http.Cookie:
		for i := j; i < len(cookiesTyp); i++ {
			if cookiesTyp[i] == nil {
				continue
			}
			writeCookie(w, cookiesTyp[i])
		}
	case []*Cookie:
		for i := j; i < len(cookiesTyp); i++ {
			if cookiesTyp[i] == nil {
				continue
			}
			writeCookie(w, &cookiesTyp[i].Cookie)
		}
	}
}

type netscapeBool bool

func (b netscapeBool) String() string {
	if b {
		return `TRUE`
	}
	return `FALSE`
}
