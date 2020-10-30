package kooky

import (
	"fmt"
	"io"
	"strings"
)

const httpOnlyPrefix = `#HttpOnly_`

// ExportCookies() export "cookies" in the Netscape format.
//
// curl, wget, ... use this format.
func ExportCookies(w io.Writer, cookies []*Cookie) {
	var isInit bool
	for _, cookie := range cookies {
		if cookie == nil {
			continue
		}
		if !isInit {
			fmt.Fprint(w, "# HTTP Cookie File\n\n")
			isInit = true
		}

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
}

type netscapeBool bool

func (b netscapeBool) String() string {
	if b {
		return `TRUE`
	}
	return `FALSE`
}
