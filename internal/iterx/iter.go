package iterx

import (
	"context"
	"errors"

	"github.com/browserutils/kooky"
)

func CookieFilterYield(ctx context.Context, cookie *kooky.Cookie, errCookie error, yield func(*kooky.Cookie, error) bool, filters ...kooky.Filter) bool {
	if errCookie != nil {
		if errors.Is(errCookie, ErrYieldEnd) {
			return false
		}
		return yield(nil, errCookie)
	}
	if kooky.FilterCookie(ctx, cookie, filters...) {
		return yield(cookie, nil)
	}
	return true
}

// NewLazyCookieFilterYielder returns a yield helper that defers cookie value
// retrieval (e.g. decryption) until non-value filters have passed.
// When splitFilters is true, filters are split into value and non-value filters
// so that expensive operations like decryption are skipped for filtered-out cookies.
func NewLazyCookieFilterYielder(splitFilters bool, filters ...kooky.Filter) func(_ context.Context, yield func(*kooky.Cookie, error) bool, _ *kooky.Cookie, errCookie error, valRetriever func(*kooky.Cookie) error) bool {
	var valueFilters, nonValueFilters []kooky.Filter
	if splitFilters {
		for _, filter := range filters {
			if _, ok := filter.(kooky.ValueFilterFunc); ok {
				valueFilters = append(valueFilters, filter)
			} else {
				// these non-value filters can be used for prefiltering before value decryption
				nonValueFilters = append(nonValueFilters, filter)
			}
		}
	}
	return func(ctx context.Context, yield func(*kooky.Cookie, error) bool, cookie *kooky.Cookie, errCookie error, valRetriever func(*kooky.Cookie) error) bool {
		if errCookie != nil {
			return yield(nil, errCookie)
		}
		if cookie == nil {
			return true
		}
		if ctx.Err() != nil {
			return false
		}
		retr := func(cookie *kooky.Cookie) bool {
			err := valRetriever(cookie)
			return err == nil || (yield(nil, err) && false)
		}
		if valRetriever != nil {
			if kooky.FilterCookie(ctx, cookie, nonValueFilters...) &&
				retr(cookie) &&
				kooky.FilterCookie(ctx, cookie, valueFilters...) {
				return yield(cookie, nil)
			}
		} else {
			if kooky.FilterCookie(ctx, cookie, filters...) {
				return yield(cookie, nil)
			}
		}
		return true
	}
}

func ErrCookieSeq(e error) kooky.CookieSeq {
	return func(yield func(*kooky.Cookie, error) bool) { yield(nil, e) }
}

// this error should not surface to the library user
var ErrYieldEnd = errors.New(`yield end`)
