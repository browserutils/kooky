package kooky

import (
	"net/http"
	"time"
)

// TODO(zellyn): figure out what to do with quoted values, like the "bcookie" cookie
// from slideshare.net

// Cookie is the struct returned by functions in this package. Similar to http.Cookie.
type Cookie struct {
	http.Cookie
	Creation  time.Time
	Container string
}
