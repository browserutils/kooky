package kooky

import (
	"net/http"
	"time"
)

// Cookie is an http.Cookie augmented with information obtained through the scraping process.
type Cookie struct {
	http.Cookie
	Creation  time.Time
	Container string
}
