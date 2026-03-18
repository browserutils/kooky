package firefox

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/iterx"
)

// SessionCookieStore reads session cookies from Firefox session store files
// (recovery.jsonlz4, recovery.baklz4, sessionstore.jsonlz4).
type SessionCookieStore struct {
	cookies.DefaultCookieStore
	Containers     map[int]string
	sessionCookies []sessionStoreCookie
}

var _ cookies.CookieStore = (*SessionCookieStore)(nil)

func (s *SessionCookieStore) Open() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.sessionCookies != nil {
		return nil
	}

	f, err := os.Open(s.FileNameStr)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := decompressMozLz4(f)
	if err != nil {
		return err
	}

	var store sessionStoreData
	if err := json.Unmarshal(data, &store); err != nil {
		return err
	}

	s.sessionCookies = store.Cookies
	return nil
}

func (s *SessionCookieStore) Close() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	s.sessionCookies = nil
	return nil
}

func (s *SessionCookieStore) TraverseCookies(filters ...kooky.Filter) kooky.CookieSeq {
	if s == nil {
		return iterx.ErrCookieSeq(errors.New(`cookie store is nil`))
	}
	if err := s.Open(); err != nil {
		return iterx.ErrCookieSeq(err)
	}

	return func(yield func(*kooky.Cookie, error) bool) {
		for _, sc := range s.sessionCookies {
			cookie := &kooky.Cookie{}
			cookie.Name = sc.Name
			cookie.Value = sc.Value
			cookie.Domain = sc.Host
			cookie.Path = sc.Path
			cookie.Secure = sc.Secure
			cookie.HttpOnly = sc.HTTPOnly
			// session cookies: zero expiry
			cookie.Expires = time.Time{}
			cookie.Browser = s

			switch sc.SameSite {
			case 1:
				cookie.SameSite = http.SameSiteNoneMode
			case 2:
				cookie.SameSite = http.SameSiteLaxMode
			case 3:
				cookie.SameSite = http.SameSiteStrictMode
			default:
				cookie.SameSite = http.SameSiteDefaultMode
			}

			// container
			if s.Containers != nil && sc.OriginAttributes.UserContextID > 0 {
				ucidStr := strconv.Itoa(sc.OriginAttributes.UserContextID)
				cookie.Container = ucidStr
				if contName, ok := s.Containers[sc.OriginAttributes.UserContextID]; ok && len(contName) > 0 {
					cookie.Container += `|` + contName
				}
			}

			if !iterx.CookieFilterYield(context.Background(), cookie, nil, yield, filters...) {
				return
			}
		}
	}
}

type sessionStoreData struct {
	Cookies []sessionStoreCookie `json:"cookies"`
}

type sessionStoreCookie struct {
	Host             string                    `json:"host"`
	Name             string                    `json:"name"`
	Value            string                    `json:"value"`
	Path             string                    `json:"path"`
	Secure           bool                      `json:"secure"`
	HTTPOnly         bool                      `json:"httponly"`
	Expiry           int64                     `json:"expiry"`
	SameSite         int                       `json:"sameSite"`
	SchemeMap        int                       `json:"schemeMap"`
	OriginAttributes sessionStoreOriginAttribs `json:"originAttributes"`
}

type sessionStoreOriginAttribs struct {
	UserContextID int `json:"userContextId"`
}
