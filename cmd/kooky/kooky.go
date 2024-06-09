package main

import (
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

// TODO: "kooky -b firefox -o /dev/stdout | head" hangs
