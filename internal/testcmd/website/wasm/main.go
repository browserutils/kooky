//go:build js

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"syscall/js"

	"github.com/browserutils/kooky/internal/website"
)

var console = js.Global().Get("console")

func logf(format string, args ...any) {
	console.Call("log", fmt.Sprintf(format, args...))
}

func main() {
	global := js.Global()
	doc := global.Get("document")
	setStatus := func(msg string) {
		doc.Call("getElementById", "status").Set("textContent", msg)
	}

	location := global.Get("location")
	logf("location.hostname = %q", location.Get("hostname").String())
	logf("location.pathname = %q", location.Get("pathname").String())
	logf("location.protocol = %q", location.Get("protocol").String())
	logf("location.href = %q", location.Get("href").String())
	logf("cookieStore defined = %t", !global.Get("cookieStore").IsUndefined() && !global.Get("cookieStore").IsNull())
	logf("document.cookie = %q", doc.Get("cookie").String())

	var api string
	var cookies []website.Cookie
	for c, err := range website.TraverseCookies(nil, &api) {
		if err != nil {
			logf("TraverseCookies error: %v", err)
			setStatus(fmt.Sprintf("error: %v", err))
			return
		}
		cookies = append(cookies, c)
	}

	logf("api used: %s", api)
	logf("cookie count: %d", len(cookies))

	for i, c := range cookies {
		logf("cookie[%d]: Name=%q Value=%q Domain=%q Path=%q Expires=%v Secure=%t SameSite=%d Partitioned=%t",
			i, c.Name, c.Value, c.Domain, c.Path, c.Expires, c.Secure, c.SameSite, c.Partitioned)
	}

	// build report and expose as JS global for export button and auto-send
	reportCookies := make([]website.ReportCookie, len(cookies))
	for i, c := range cookies {
		var expires string
		if !c.Expires.IsZero() {
			expires = c.Expires.UTC().Format(website.ExpiresFormat)
		}
		d := c.Derived
		reportCookies[i] = website.ReportCookie{
			Name:        c.Name,
			Value:       c.Value,
			Domain:      c.Domain,
			Path:        c.Path,
			Expires:     expires,
			Secure:      c.Secure,
			SameSite:    website.SameSiteString(c.SameSite),
			Partitioned: c.Partitioned,
			Derived:     &d,
		}
	}
	report := website.CookieReport{
		UserAgent: global.Get("navigator").Get("userAgent").String(),
		URL:       location.Get("href").String(),
		API:       api,
		Cookies:   reportCookies,
	}
	reportJSON, _ := json.MarshalIndent(report, "", "  ")
	reportStr := string(reportJSON)
	global.Set("_kooky_report", reportStr)

	// send report to server
	go func() {
		resp, err := http.Post(location.Get("origin").String()+"/api/report", "application/json", strings.NewReader(reportStr))
		if err != nil {
			logf("report send error: %v", err)
			return
		}
		resp.Body.Close()
	}()

	setStatus(fmt.Sprintf("%d cookie(s) found (via %s)", len(cookies), api))
	doc.Call("getElementById", "filters").Get("style").Set("display", "flex")

	// render table
	app := doc.Call("getElementById", "app")
	table := doc.Call("createElement", "table")
	table.Set("id", "cookie-table")

	thead := doc.Call("createElement", "thead")
	hrow := doc.Call("createElement", "tr")
	for _, col := range []string{"Name", "Value", "Domain", "Path", "Expires", "Secure", "SameSite", "Partitioned"} {
		th := doc.Call("createElement", "th")
		th.Set("textContent", col)
		hrow.Call("appendChild", th)
	}
	thead.Call("appendChild", hrow)
	table.Call("appendChild", thead)

	tbody := doc.Call("createElement", "tbody")
	for _, rc := range reportCookies {
		tr := doc.Call("createElement", "tr")
		addCell := func(text string) {
			td := doc.Call("createElement", "td")
			td.Set("textContent", text)
			tr.Call("appendChild", td)
		}
		addCell(rc.Name)
		addCell(rc.Value)
		addCell(rc.Domain)
		addCell(rc.Path)
		addCell(rc.Expires)
		addCell(fmt.Sprintf("%t", rc.Secure))
		addCell(rc.SameSite)
		addCell(fmt.Sprintf("%t", rc.Partitioned))
		tbody.Call("appendChild", tr)
	}
	table.Call("appendChild", tbody)
	app.Call("appendChild", table)

	// keep wasm alive
	select {}
}
