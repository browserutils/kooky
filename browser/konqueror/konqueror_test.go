package konqueror

import (
	"testing"
	"time"

	"github.com/xiazemin/kooky/internal/testutils"
)

func TestReadCookies(t *testing.T) {
	testCookiesPath, err := testutils.GetTestDataFilePath("konqueror-cookies")
	if err != nil {
		t.Fatalf("Failed to load test data file")
	}

	cookies, err := ReadCookies(testCookiesPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(cookies) != 16 {
		t.Fatalf("got %d cookies, but expected 16", len(cookies))
	}

	// timezone
	tz := time.Local

	c := cookies[1]
	if c.Domain != ".google.de" {
		t.Errorf("c.Domain=%q", c.Domain)
	}
	if c.Name != "CONSENT" {
		t.Errorf("c.Name=%q", c.Name)
	}
	if c.Path != "/" {
		t.Errorf("c.Path=%q", c.Path)
	}
	if !c.Expires.Equal(time.Date(2038, 1, 10, 8, 59, 59, 0, tz)) {
		t.Errorf("c.Expires=%q", c.Expires)
	}
	if !c.Secure {
		t.Error("c.Secure expected true")
	}
	if c.HttpOnly {
		t.Error("c.HttpOnly expected false")
	}
	// Creation fields unused by parser
	if c.Value != "some-value" {
		t.Errorf("c.Value=%q", c.Value)
	}

	// test bit flags again
	c = cookies[11]
	if c.Secure {
		t.Error("c.Secure expected false")
	}
	if !c.HttpOnly {
		t.Error("c.HttpOnly expected true")
	}

	// test cookie with empty Domain field
	c = cookies[13]
	if c.Domain != "www.amazon.de" {
		t.Errorf("c.Domain=%q", c.Domain)
	}
}
