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

	for i := j; i < len(cookies); i++ {
		if cookies[i] == nil {
			continue
		}

		var domain string
		if cookies[i].HttpOnly {
			domain = httpOnlyPrefix
		}
		domain += cookies[i].Domain

		fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
			domain,
			netscapeBool(strings.HasPrefix(cookies[i].Domain, `.`)),
			cookies[i].Path,
			netscapeBool(cookies[i].Secure),
			cookies[i].Expires.Unix(),
			cookies[i].Name,
			cookies[i].Value,
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
