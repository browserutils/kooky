package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all"

	"github.com/spf13/pflag"
)

func main() {
	browser := pflag.StringP(`browser`, `b`, ``, `browser filter`)
	profile := pflag.StringP(`profile`, `p`, ``, `profile filter`)
	defaultProfile := pflag.BoolP(`default-profile`, `q`, false, `only default profile(s)`)
	showExpired := pflag.BoolP(`expired`, `e`, false, `show expired cookies`)
	domain := pflag.StringP(`domain`, `d`, ``, `cookie domain filter (partial)`)
	name := pflag.StringP(`name`, `n`, ``, `cookie name filter (exact)`)
	export := pflag.StringP(`export`, `o`, ``, `export cookies in netscape format`)
	jsonFormat := pflag.BoolP(`jsonl`, `j`, false, `JSON Lines output format`)
	pflag.Parse()

	cookieStores := kooky.FindAllCookieStores()

	var cookiesExport []*kooky.Cookie // for netscape export

	var f io.Writer         // for netscape export
	var w *tabwriter.Writer // for printing
	if export != nil && len(*export) > 0 {
		if *export == `-` {
			f = os.Stdout
		} else {
			fl, err := os.OpenFile(*export, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				log.Fatalln(err)
			}
			defer fl.Close()
			f = fl
		}
	} else {
		w = tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	}
	trimLen := 45
	for _, store := range cookieStores {
		defer store.Close()

		// cookie store filters
		if browser != nil && len(*browser) > 0 && store.Browser() != *browser {
			continue
		}
		if profile != nil && len(*profile) > 0 && store.Profile() != *profile {
			continue
		}
		if defaultProfile != nil && *defaultProfile && !store.IsDefaultProfile() {
			continue
		}

		// cookie filters
		var filters []kooky.Filter
		if showExpired == nil || !*showExpired {
			filters = append(filters, kooky.Valid)
		}
		if domain != nil && len(*domain) > 0 {
			filters = append(filters, kooky.DomainContains(*domain))
		}
		if name != nil && len(*name) > 0 {
			filters = append(filters, kooky.Name(*name))
		}

		cookies, _ := store.ReadCookies(filters...)

		if export != nil && len(*export) > 0 {
			cookiesExport = append(cookiesExport, cookies...)
		} else {
			for _, cookie := range cookies {
				if jsonFormat != nil && *jsonFormat {
					c := &jsonCookieExtra{
						Cookie: cookie,
						jsonCookieExtraFields: jsonCookieExtraFields{
							Browser:  store.Browser(),
							Profile:  store.Profile(),
							FilePath: store.FilePath(),
						},
					}
					b, err := json.Marshal(c)
					if err != nil {
						log.Fatalln(err)
					}
					fmt.Fprintf(w, "%s\n", b)
				} else {
					container := cookie.Container
					if len(container) > 0 {
						container = ` [` + container + `]`
					}
					fmt.Fprintf(
						w,
						"%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						store.Browser(),
						store.Profile(),
						container,
						trimStr(store.FilePath(), trimLen),
						trimStr(cookie.Domain, trimLen),
						trimStr(cookie.Name, trimLen),
						// be careful about raw bytes
						trimStr(strings.Trim(fmt.Sprintf(`%q`, cookie.Value), `"`), trimLen),
						cookie.Expires.Format(`2006.01.02 15:04:05`),
					)
				}
			}
		}
	}
	if export != nil && len(*export) > 0 {
		kooky.ExportCookies(f, cookiesExport)
	} else {
		w.Flush()
	}
}

func trimStr(str string, length int) string {
	if len(str) <= length {
		return str
	}
	if length > 0 {
		return str[:length-1] + "\u2026" // "..."
	}
	return str[:length]
}

type jsonCookieExtra struct {
	*kooky.Cookie
	jsonCookieExtraFields
}

type jsonCookieExtraFields struct {
	// separated for easier json marshaling
	Browser  string `json:"browser,omitempty"`
	Profile  string `json:"profile,omitempty"`
	FilePath string `json:"file_path,omitempty"`
}

func (c *jsonCookieExtra) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte(`null`), nil
	}
	b := &strings.Builder{}
	b.WriteByte('{')
	var hasCookieBytes bool
	if c.Cookie != nil {
		bc, err := c.Cookie.MarshalJSON()
		if err != nil {
			return nil, err
		}
		hasCookieBytes = len(bc) > 2
		if hasCookieBytes {
			b.Write(bc[1 : len(bc)-1])
		}
	}
	be, err := json.Marshal(c.jsonCookieExtraFields)
	if err != nil || len(be) <= 2 {
		b.WriteByte('}')
		return []byte(b.String()), nil
	}
	if hasCookieBytes {
		b.WriteByte(',')
	}
	b.Write(append(be[1:len(be)-1], '}'))
	return []byte(b.String()), nil
}

// TODO: "kooky -b firefox -o /dev/stdout | head" hangs
