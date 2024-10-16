package kooky

import (
	"encoding/json"
	"net/http"
	"time"
)

// Cookie is an http.Cookie augmented with information obtained through the scraping process.
type Cookie struct {
	http.Cookie
	Creation  time.Time
	Container string
}

// adjustments to the json marshaling to allow dates with more than 4 year digits
// https://github.com/golang/go/issues/4556
// https://github.com/golang/go/issues/54580
// encoding/json/v2 "format"(?) might make this unnecessary

func (c *Cookie) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte(`null`), nil
	}
	c2 := &struct {
		// net/http.Cookie
		Name        string        `json:"name"`
		Value       string        `json:"value"`
		Quoted      bool          `json:"quoted"`
		Path        string        `json:"path"`
		Domain      string        `json:"domain"`
		Expires     jsonTime      `json:"expires"`
		RawExpires  string        `json:"raw_expires,omitempty"`
		MaxAge      int           `json:"max_age"`
		Secure      bool          `json:"secure"`
		HttpOnly    bool          `json:"http_only"`
		SameSite    http.SameSite `json:"same_site"`
		Partitioned bool          `json:"partitioned"`
		Raw         string        `json:"raw,omitempty"`
		Unparsed    []string      `json:"unparsed,omitempty"`
		// extra fields
		Creation  jsonTime `json:"creation"`
		Container string   `json:"container,omitempty"`
	}{
		Name:        c.Cookie.Name,
		Value:       c.Cookie.Value,
		Quoted:      c.Cookie.Quoted,
		Path:        c.Cookie.Path,
		Domain:      c.Cookie.Domain,
		Expires:     jsonTime{c.Cookie.Expires},
		RawExpires:  c.Cookie.RawExpires,
		MaxAge:      c.Cookie.MaxAge,
		Secure:      c.Cookie.Secure,
		HttpOnly:    c.Cookie.HttpOnly,
		SameSite:    c.Cookie.SameSite,
		Partitioned: c.Cookie.Partitioned,
		Raw:         c.Cookie.Raw,
		Unparsed:    c.Cookie.Unparsed,
		Creation:    jsonTime{c.Creation},
		Container:   c.Container,
	}
	return json.Marshal(c2)
}

type jsonTime struct{ time.Time }

// MarshalJSON implements the [json.Marshaler] interface.
// The time is a quoted string in the RFC 3339 format with sub-second precision.
// the timestamp might be represented as invalid RFC 3339 if necessary (year with more than 4 digits).
func (t jsonTime) MarshalJSON() ([]byte, error) {
	if b, err := t.Time.MarshalJSON(); err == nil {
		return b, nil
	}
	return []byte(t.Time.Format(`"` + time.RFC3339 + `"`)), nil
}
