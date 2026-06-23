package website

import (
	"net/http"
	"strings"
)

// ExpiresFormat is the time layout used for cookie expiry in reports.
const ExpiresFormat = "2006-01-02T15:04:05Z"

func SameSiteString(s http.SameSite) string {
	switch s {
	case http.SameSiteStrictMode:
		return "Strict"
	case http.SameSiteLaxMode:
		return "Lax"
	case http.SameSiteNoneMode:
		return "None"
	default:
		return ""
	}
}

func ParseSameSite(s string) http.SameSite {
	switch strings.ToLower(s) {
	case "strict":
		return http.SameSiteStrictMode
	case "lax":
		return http.SameSiteLaxMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return 0
	}
}

// CookieReport is the JSON structure sent by the browser.
type CookieReport struct {
	UserAgent string         `json:"userAgent"`
	URL       string         `json:"url"`
	API       string         `json:"api"`
	Cookies   []ReportCookie `json:"cookies"`
}

type ReportCookie struct {
	Name        string         `json:"name"`
	Value       string         `json:"value"`
	Domain      string         `json:"domain"`
	Path        string         `json:"path"`
	Expires     string         `json:"expires"`
	Secure      bool           `json:"secure"`
	SameSite    string         `json:"sameSite"`
	Partitioned bool           `json:"partitioned"`
	Derived     *DerivedFields `json:"derived,omitempty"`
}
