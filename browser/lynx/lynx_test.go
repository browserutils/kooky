package lynx

import (
	"testing"
	"time"

	"github.com/xiazemin/kooky/internal/testutils"
)

func TestReadCookies(t *testing.T) {
	testCookiesPath, err := testutils.GetTestDataFilePath("lynx-.lynx_cookies")
	if err != nil {
		t.Fatalf("Failed to load test data file")
	}

	cookies, err := ReadCookies(testCookiesPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(cookies) != 1 {
		t.Fatalf("got %d cookies, but expected 1", len(cookies))
	}

	// timezone
	tz := time.Local

	c := cookies[0]
	if c.Domain != "google.de" {
		t.Errorf("c.Domain=%q", c.Domain)
	}
	if c.Name != "NID" {
		t.Errorf("c.Name=%q", c.Name)
	}
	if c.Path != "/" {
		t.Errorf("c.Path=%q", c.Path)
	}
	if !c.Expires.Equal(time.Date(2021, 4, 16, 10, 38, 33, 0, tz)) {
		t.Errorf("c.Expires=%q", c.Expires)
	}
	if c.Secure {
		t.Error("c.Secure expected false")
	}
	if c.HttpOnly {
		t.Error("c.HttpOnly expected false")
	}
	// Creation fields unused by parser
	if c.Value != "204=blabla" {
		t.Errorf("c.Value=%q", c.Value)
	}
}
