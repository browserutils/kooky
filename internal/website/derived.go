package website

import "github.com/browserutils/kooky"

// Cookie pairs a kooky.Cookie with its derivation metadata.
type Cookie struct {
	*kooky.Cookie
	Derived DerivedFields
}

// DerivedFields tracks which cookie fields were inferred from
// window.location rather than provided by the browser's cookie API.
type DerivedFields struct {
	// Domain was not returned by the API; filled from location.hostname.
	Domain bool `json:"domain,omitempty"`
	// Path was not returned by the API; filled from location.pathname.
	// This is the current page path and may be a subset of the actual
	// cookie path — we cannot know at which parent path level the cookie
	// was originally set.
	Path bool `json:"path,omitempty"`
	// Secure was not returned by the API; inferred from location.protocol.
	Secure bool `json:"secure,omitempty"`
}
