package ladybird

import (
	"context"
	"errors"
	"time"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/iterx"
	"github.com/browserutils/kooky/internal/utils"
)

func ReadCookies(ctx context.Context, filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	return cookies.SingleRead(cookieStore, filename, filters...).ReadAllCookies(ctx)
}

func TraverseCookies(filename string, filters ...kooky.Filter) kooky.CookieSeq {
	return cookies.SingleRead(cookieStore, filename, filters...)
}

func (s *ladybirdCookieStore) TraverseCookies(filters ...kooky.Filter) kooky.CookieSeq {
	if s == nil {
		return iterx.ErrCookieSeq(errors.New(`cookie store is nil`))
	}
	if err := s.Open(); err != nil {
		return iterx.ErrCookieSeq(err)
	} else if s.Database == nil {
		return iterx.ErrCookieSeq(errors.New(`database is nil`))
	}

	visitor := func(yield func(*kooky.Cookie, error) bool) func(rowId *int64, row utils.TableRow) error {
		return func(rowId *int64, row utils.TableRow) error {
			cookie := kooky.Cookie{}
			var err error

			cookie.Name, err = row.String(`name`)
			if err != nil {
				return err
			}

			cookie.Value, err = row.String(`value`)
			if err != nil {
				return err
			}

			cookie.Domain, err = row.String(`domain`)
			if err != nil {
				return err
			}

			cookie.Path, err = row.String(`path`)
			if err != nil {
				return err
			}

			// Ladybird stores timestamps as Unix milliseconds
			expiryMs, err := utils.ValueOrFallback[int64](row, `expiry_time`, 0, true)
			if err != nil {
				return err
			}
			cookie.Expires = time.UnixMilli(expiryMs)

			creationMs, err := utils.ValueOrFallback[int64](row, `creation_time`, 0, true)
			if err == nil && creationMs > 0 {
				cookie.Creation = time.UnixMilli(creationMs)
			}

			cookie.Secure, err = row.Bool(`secure`)
			if err != nil {
				return err
			}

			cookie.HttpOnly, err = row.Bool(`http_only`)
			if err != nil {
				return err
			}

			cookie.Browser = s

			if !iterx.CookieFilterYield(context.Background(), &cookie, nil, yield, filters...) {
				return iterx.ErrYieldEnd
			}

			return nil
		}
	}
	seq := func(yield func(*kooky.Cookie, error) bool) {
		err := utils.VisitTableRows(s.Database, `Cookies`, map[string]string{}, visitor(yield))
		if err != nil && !errors.Is(err, iterx.ErrYieldEnd) {
			yield(nil, err)
		}
	}

	return seq
}

// CookieStore has to be closed with CookieStore.Close() after use.
func CookieStore(filename string, filters ...kooky.Filter) (kooky.CookieStore, error) {
	return cookieStore(filename, filters...)
}

func cookieStore(filename string, filters ...kooky.Filter) (*cookies.CookieJar, error) {
	s := &ladybirdCookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `ladybird`

	return cookies.NewCookieJar(s, filters...), nil
}
