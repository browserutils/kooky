package firefox

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/iterx"
	"github.com/browserutils/kooky/internal/utils"
)

func (s *CookieStore) TraverseCookies(filters ...kooky.Filter) kooky.CookieSeq {
	if s == nil {
		return iterx.ErrCookieSeq(errors.New(`cookie store is nil`))
	}
	if err := s.Open(); err != nil {
		return iterx.ErrCookieSeq(err)
	} else if s.Database == nil {
		return iterx.ErrCookieSeq(errors.New(`database is nil`))
	}

	_ = s.initContainersMap()

	visitor := func(yield func(*kooky.Cookie, error) bool) func(rowId *int64, row utils.TableRow) error {
		return func(rowId *int64, row utils.TableRow) error {
			cookie := kooky.Cookie{}
			var err error

			// Name
			cookie.Name, err = row.String(`name`)
			if err != nil {
				return err
			}

			// Value
			cookie.Value, err = row.String(`value`)
			if err != nil {
				return err
			}

			// Domain
			if baseDomain := row.ValueOrFallback(`baseDomain`, nil); baseDomain == nil {
				if host, err := row.String(`host`); err != nil {
					return err
				} else {
					cookie.Domain = host
				}
			} else {
				// handle databases prior v78 ESR
				var ok bool
				cookie.Domain, ok = baseDomain.(string)
				if !ok {
					return fmt.Errorf("got unexpected value for baseDomain %v (type %[1]T)", baseDomain)
				}
			}

			// Path
			cookie.Path, err = row.String(`path`)
			if err != nil {
				return err
			}

			// Expires
			if expiry, err := row.Int64(`expiry`); err == nil {
				cookie.Expires = time.Unix(expiry, 0)
			} else {
				return err
			}

			// Creation
			if creationTime, err := row.Int64(`creationTime`); err == nil {
				cookie.Creation = time.UnixMicro(creationTime)
			} else {
				return err
			}

			// Secure
			cookie.Secure, err = row.Bool(`isSecure`)
			if err != nil {
				return err
			}

			// HttpOnly
			cookie.HttpOnly, err = row.Bool(`isHttpOnly`)
			if err != nil {
				return err
			}

			// Container
			if s.Containers != nil {
				ucidStr, _ := row.String(`originAttributes`)
				prefixContextID := `^userContextId=`
				if len(ucidStr) > 0 && strings.HasPrefix(ucidStr, prefixContextID) {
					ucidStr = strings.TrimPrefix(ucidStr, prefixContextID)
					cookie.Container = ucidStr
					ucid, err := strconv.Atoi(ucidStr)
					if err == nil {
						contName, okContName := s.Containers[ucid]
						if okContName && len(contName) > 0 {
							cookie.Container += `|` + contName
						}
					}
				}
			}

			cookie.Browser = s

			if !iterx.CookieFilterYield(context.Background(), &cookie, nil, yield, filters...) {
				return iterx.ErrYieldEnd
			}
			return nil
		}
	}
	seq := func(yield func(*kooky.Cookie, error) bool) {
		err := utils.VisitTableRows(s.Database, `moz_cookies`, map[string]string{}, visitor(yield))
		if !errors.Is(err, iterx.ErrYieldEnd) {
			yield(nil, err)
		}
	}

	return seq
}
