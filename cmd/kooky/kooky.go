package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/zellyn/kooky"
	_ "github.com/zellyn/kooky/allbrowsers"

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
	var w *tabwriter.Writer
	if export == nil || len(*export) < 1 {
		w = tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	}
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
			if *export == `-` {
				kooky.ExportCookies(os.Stdout, cookies)
			} else {
				f, err := os.OpenFile(*export, os.O_RDWR|os.O_CREATE, 0644)
				if err != nil {
					log.Fatalln(err)
				}
				defer f.Close()
				kooky.ExportCookies(f, cookies)
			}
		} else {
			trimLen := 45
			for _, cookie := range cookies {
				fmt.Fprintf(
					w,
					"%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					store.Browser(),
					store.Profile(),
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
	if export == nil || len(*export) < 1 {
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
