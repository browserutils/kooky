//go:build !js

package website

import (
	"errors"
	"iter"

	"github.com/browserutils/kooky"
)

var errUnsupportedPlatform = errors.New(`website: cookie access requires js platform`)

// TraverseCookies is not available on non-js platforms.
func TraverseCookies(kooky.BrowserInfo, *string, ...kooky.Filter) iter.Seq2[Cookie, error] {
	return func(yield func(Cookie, error) bool) {
		yield(Cookie{}, errUnsupportedPlatform)
	}
}
