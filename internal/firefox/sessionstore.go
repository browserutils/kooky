package firefox

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/iterx"
)

// sessionStoreFiles lists session store files ordered by freshness:
// runtime files first (most current), then shutdown files.
var sessionStoreFiles = []string{
	filepath.Join(`sessionstore-backups`, `recovery.jsonlz4`),
	filepath.Join(`sessionstore-backups`, `recovery.baklz4`),
	`sessionstore.jsonlz4`,
	filepath.Join(`sessionstore-backups`, `previous.jsonlz4`),
}

// SessionCookieStore reads session cookies from Firefox session store files.
// FileNameStr is the profile directory; the store resolves the actual file on access.
type SessionCookieStore struct {
	cookies.DefaultCookieStore
	Containers     map[int]string
	containersErr  error
	profileDir     string
	resolvedPath   string
	sessionCookies []sessionStoreCookie
}

var _ cookies.CookieStore = (*SessionCookieStore)(nil)

func (s *SessionCookieStore) resolveFile() string {
	dir := s.profileDir
	if dir == `` {
		dir = s.FileNameStr
	}
	for _, name := range sessionStoreFiles {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			s.resolvedPath = path
			return path
		}
	}
	return ``
}

func (s *SessionCookieStore) FilePath() string {
	if s == nil {
		return ``
	}
	if s.resolvedPath != `` {
		return s.resolvedPath
	}
	return s.resolveFile()
}

func (s *SessionCookieStore) Open() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.profileDir == `` {
		s.profileDir = s.FileNameStr
	}

	s.initSessionContainersMap()

	s.resolvedPath = ``
	s.sessionCookies = nil

	var errs []error
	for _, name := range sessionStoreFiles {
		path := filepath.Join(s.profileDir, name)
		store, err := readSessionStoreFile(path)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		s.resolvedPath = path
		s.FileNameStr = path
		s.sessionCookies = store.Cookies
		return nil
	}
	return errors.Join(errs...)
}

func readSessionStoreFile(path string) (*sessionStoreData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := decompressMozLz4(f)
	if err != nil {
		return nil, err
	}

	var store sessionStoreData
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	return &store, nil
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
	// always re-read; session store files are rewritten frequently by Firefox
	if err := s.Open(); err != nil {
		return iterx.ErrCookieSeq(err)
	}

	return func(yield func(*kooky.Cookie, error) bool) {
		if s.containersErr != nil {
			if !yield(nil, s.containersErr) {
				return
			}
		}
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
			// TODO: creation time?
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
					cookie.Container = contName
				}
			}
			// CHIPS partitioned cookie
			if len(sc.OriginAttributes.PartitionKey) > 0 {
				cookie.Partitioned = true
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
	UserContextID int    `json:"userContextId"`
	PartitionKey  string `json:"partitionKey"`
}
