package iterx

import (
	"cmp"
	"context"
	"errors"

	"github.com/browserutils/kooky"
)

func CookieFilterYield(ctx context.Context, cookie *kooky.Cookie, errCookie error, yield func(*kooky.Cookie, error) bool, filters ...kooky.Filter) bool {
	ret := true
	if errCookie != nil {
		if errors.Is(errCookie, ErrYieldEnd) {
			return false
		}
		ret = yield(nil, errCookie)
	}
	if kooky.FilterCookie(ctx, cookie, filters...) {
		ret = yield(cookie, nil)
	}
	return ret
}

func NewCookieFilterYielder(splitFilters bool, filters ...kooky.Filter) func(_ context.Context, yield func(*kooky.Cookie, error) bool, _ *kooky.Cookie, errCookie error, valRetriever func(*kooky.Cookie) error) bool {
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
		retr := func(cookie *kooky.Cookie) bool {
			err := valRetriever(cookie)
			return err == nil || (yield(nil, err) && false)
		}
		done := make(chan struct{})
		var ret, ctxDone bool
		go func() {
			defer func() { done <- struct{}{} }()
			if valRetriever != nil {
				ret = !cmp.Or(
					!kooky.FilterCookie(ctx, cookie, nonValueFilters...),
					!retr(cookie),
					!kooky.FilterCookie(ctx, cookie, valueFilters...),
				)
			} else {
				ret = kooky.FilterCookie(ctx, cookie, filters...)
			}
			ret = ret && yield(cookie, nil)
		}()
		select {
		case <-ctx.Done():
			ctxDone = true
			<-done // wait for current yield to finish
		case <-done:
		}
		return ret && !ctxDone
	}
}

func ErrCookieSeq(e error) kooky.CookieSeq {
	return func(yield func(*kooky.Cookie, error) bool) { yield(nil, e) }
}

// this error should not surface to the library user
var ErrYieldEnd = errors.New(`yield end`)
