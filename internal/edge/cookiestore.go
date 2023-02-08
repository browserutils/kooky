package edge

import (
	"github.com/zellyn/kooky/internal/chrome"
	"github.com/zellyn/kooky/internal/cookies"
)

type CookieStore struct {
	chrome.CookieStore
}

func (s *CookieStore) Open() error {
	return s.CookieStore.Open()
}

func (s *CookieStore) Close() error {
	return s.CookieStore.Close()
}

var _ cookies.CookieStore = (*CookieStore)(nil)

// returns the previous password for later restoration
// used in tests
func (s *CookieStore) SetKeyringPassword(password []byte) []byte {
	return s.CookieStore.SetKeyringPassword(password)
}
