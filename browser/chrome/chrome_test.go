package chrome

import (
	"testing"
	"time"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/chrome"
	"github.com/browserutils/kooky/internal/testutils"
)

// d18f6247db68045dfbab126d814baf2cf1512141391
func TestReadCookies(t *testing.T) {
	testCookiesPath, err := testutils.GetTestDataFilePath("chrome-macos-cookie-db.sqlite") // this test file was created on macos
	if err != nil {
		t.Fatalf("Failed to load test data file")
	}

	s := &chrome.CookieStore{}
	s.FileNameStr = testCookiesPath

	defer s.Close()
	// Prevent reading the password from the OS Keyring.
	oldPassword := s.SetKeyringPassword([]byte("ChromeSafeStoragePasswrd"))
	defer s.SetKeyringPassword(oldPassword)

	cookies, err := s.ReadCookies()
	if err != nil {
		t.Fatal(err)
	}

	domain := "news.ycombinator.com"
	name := "user"

	cookies = kooky.FilterCookies(cookies, kooky.Domain(domain), kooky.Name(name))
	if len(cookies) == 0 {
		t.Fatalf("Found no cookies with domain=%q, name=%q", domain, name)
	}
	cookie := cookies[0]

	wantValue := "zellyn&p2EXEjsXVNPxXcrZiK8DoezI4Erqt0vA"
	if cookie.Value != wantValue {
		t.Errorf("Want cookie value %q; got %q", wantValue, cookie.Value)
	}

	wantExpires := time.Date(2038, 01, 17, 19, 14, 07, 876554000, time.UTC)
	if !cookie.Expires.Equal(wantExpires) {
		t.Errorf("Want cookie.Expires=%v; got %v", wantExpires, cookie.Expires)
	}

	wantCreation := time.Date(2017, 12, 1, 15, 16, 48, 876554000, time.UTC)
	if !cookie.Creation.Equal(wantCreation) {
		t.Errorf("Want cookie.Creation=%v; got %v", wantCreation, cookie.Creation)
	}
}
