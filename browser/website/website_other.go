//go:build !js

package website

import (
	"context"
	"errors"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/iterx"
)

var errUnsupportedPlatform = errors.New(`website: cookie access requires js or wasm platform`)

func ReadCookies(_ context.Context, _ ...kooky.Filter) ([]*kooky.Cookie, error) {
	return nil, errUnsupportedPlatform
}

func TraverseCookies(_ ...kooky.Filter) kooky.CookieSeq {
	return iterx.ErrCookieSeq(errUnsupportedPlatform)
}

// CookieStore has to be closed with CookieStore.Close() after use.
func CookieStore(_ ...kooky.Filter) (kooky.CookieStore, error) {
	return nil, errUnsupportedPlatform
}
