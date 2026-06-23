package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/browserutils/kooky/internal/website"
)

func handleReport(sentCookies []*http.Cookie) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var report website.CookieReport
		if err := json.Unmarshal(body, &report); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		printReport(report, sentCookies)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}
}

func printReport(report website.CookieReport, sent []*http.Cookie) {
	// derive expected defaults from the browser's URL
	u, _ := url.Parse(report.URL)
	isHTTPS := u != nil && u.Scheme == "https"
	hostname := ""
	if u != nil {
		hostname = u.Hostname()
	}

	sep := strings.Repeat("─", 72)
	fmt.Println(sep)
	fmt.Printf("Browser:  %s\n", report.UserAgent)
	fmt.Printf("URL:      %s\n", report.URL)
	fmt.Printf("API:      %s\n", report.API)
	fmt.Printf("Received: %d cookie(s)  Sent: %d cookie(s)\n", len(report.Cookies), len(sent))
	fmt.Println(sep)

	// index received cookies by name
	received := make(map[string]website.ReportCookie, len(report.Cookies))
	for _, c := range report.Cookies {
		received[c.Name] = c
	}

	nOK := 0
	for _, s := range sent {
		rc, found := received[s.Name]
		if !found {
			var reasons []string
			if s.HttpOnly {
				reasons = append(reasons, "HttpOnly=true")
			}
			if s.Secure && !isHTTPS {
				reasons = append(reasons, "Secure=true (no HTTPS)")
			}
			if s.Path != "/" {
				reasons = append(reasons, fmt.Sprintf("Path=%q", s.Path))
			}
			if len(reasons) == 0 {
				reasons = append(reasons, "unknown")
			}
			fmt.Printf("  MISSING  %s  %s\n", s.Name, strings.Join(reasons, ", "))
			continue
		}

		isDerived := rc.Derived != nil
		var diffs []string

		if rc.Value != s.Value {
			diffs = append(diffs, fmt.Sprintf("Value: %q (expected %q)", rc.Value, s.Value))
		}

		// Domain: empty in Set-Cookie means host-only; browser fills from hostname
		expectDomain := s.Domain
		if expectDomain == "" {
			expectDomain = hostname
		}
		if rc.Domain != expectDomain {
			d := fmt.Sprintf("Domain: %q (expected %q)", rc.Domain, expectDomain)
			if isDerived && rc.Derived.Domain {
				d += " [derived from location]"
			}
			diffs = append(diffs, d)
		}

		if rc.Path != s.Path {
			d := fmt.Sprintf("Path: %q (expected %q)", rc.Path, s.Path)
			if isDerived && rc.Derived.Path {
				d += " [derived from location, actual path unknown]"
			}
			diffs = append(diffs, d)
		}

		// Secure: Cookie Store API returns actual value;
		// document.cookie cannot read it, inferred from protocol.
		if rc.Secure != s.Secure {
			d := fmt.Sprintf("Secure: %t (expected %t)", rc.Secure, s.Secure)
			if isDerived && rc.Derived.Secure {
				d += " [derived from protocol]"
			}
			diffs = append(diffs, d)
		}

		// SameSite: browsers default to Lax when not set
		sentSS := website.SameSiteString(s.SameSite)
		expectSS := sentSS
		if expectSS == "" {
			expectSS = "Lax" // browser default
		}
		if rc.SameSite != "" && rc.SameSite != expectSS {
			diffs = append(diffs, fmt.Sprintf("SameSite: %q (expected %q)", rc.SameSite, expectSS))
		}

		// Expires
		if !s.Expires.IsZero() && rc.Expires != "" {
			gotExp, err := time.Parse(website.ExpiresFormat, rc.Expires)
			if err == nil {
				diff := s.Expires.UTC().Sub(gotExp.UTC())
				if diff < -2*time.Second || diff > 2*time.Second {
					diffs = append(diffs, fmt.Sprintf("Expires: %q (expected %q)", rc.Expires, s.Expires.UTC().Format(time.RFC3339)))
				}
			}
		}
		if s.Expires.IsZero() && rc.Expires != "" {
			diffs = append(diffs, fmt.Sprintf("Expires: %q (expected session)", rc.Expires))
		}

		if len(diffs) == 0 {
			nOK++
			continue
		}
		fmt.Printf("  DIFF     %s\n", s.Name)
		for _, d := range diffs {
			fmt.Printf("             %s\n", d)
		}
	}

	// unexpected cookies from the browser
	for _, rc := range report.Cookies {
		found := false
		for _, s := range sent {
			if s.Name == rc.Name {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("  EXTRA    %s\n", rc.Name)
		}
	}

	if nOK > 0 {
		fmt.Printf("  OK       %d cookie(s)\n", nOK)
	}
	fmt.Println(sep)
}

