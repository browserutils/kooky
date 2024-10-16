package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
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

	// cookie filters
	filters := []kooky.Filter{storeFilter(browser, profile, defaultProfile)}
	if showExpired == nil || !*showExpired {
		filters = append(filters, kooky.Valid)
	}
	if domain != nil && len(*domain) > 0 {
		filters = append(filters, kooky.DomainContains(*domain))
	}
	if name != nil && len(*name) > 0 {
		filters = append(filters, kooky.Name(*name))
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	seq := kooky.TraverseCookies(ctx, filters...)

	if export != nil && len(*export) > 0 {
		var f io.Writer // for netscape export
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
		seq.Export(ctx, f)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0) // for printing

	trimLen := 45
	// use channel so that tabwriter won't panic
	for cookie := range seq.Chan(ctx) {
		if jsonFormat != nil && *jsonFormat {
			b, err := json.Marshal(cookie)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Fprintf(w, "%s\n", b)
		} else {
			prCookieLine(w, cookie, trimLen)
		}
	}
}

func prCookieLine(w io.Writer, cookie *kooky.Cookie, trimLen int) {
	if cookie == nil {
		return
	}
	container := cookie.Container
	if len(container) > 0 {
		container = ` [` + container + `]`
	}
	fmt.Fprintf(
		w,
		"%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		prBrowser(cookie),
		prProfile(cookie),
		container,
		trimStr(prFilePath(cookie), trimLen),
		trimStr(cookie.Domain, trimLen),
		trimStr(cookie.Name, trimLen),
		// be careful about raw bytes
		trimStr(strings.Trim(fmt.Sprintf(`%q`, cookie.Value), `"`), trimLen),
		cookie.Expires.Format(`2006.01.02 15:04:05`),
	)
}

func prBrowser(c *kooky.Cookie) string {
	if c == nil || c.Browser == nil {
		return ``
	}
	return c.Browser.Browser()
}

func prProfile(c *kooky.Cookie) string {
	if c == nil || c.Browser == nil {
		return ``
	}
	return c.Browser.Profile()
}

func prFilePath(c *kooky.Cookie) string {
	if c == nil || c.Browser == nil {
		return ``
	}
	return c.Browser.FilePath()
}

func storeFilter(browser, profile *string, defaultProfile *bool) kooky.Filter {
	return kooky.FilterFunc(func(cookie *kooky.Cookie) bool {
		if cookie == nil || cookie.Browser == nil {
			return false
		}
		// cookie store filters
		if browser != nil && len(*browser) > 0 && cookie.Browser.Browser() != *browser {
			return false
		}
		if profile != nil && len(*profile) > 0 && cookie.Browser.Profile() != *profile {
			return false
		}
		if defaultProfile != nil && *defaultProfile && !cookie.Browser.IsDefaultProfile() {
			return false
		}
		return true
	})
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
