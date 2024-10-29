package netscape

import (
	"context"
	"testing"
	"time"

	"github.com/xiazemin/kooky/internal/testutils"
)

func TestReadCookies(t *testing.T) {
	testCookiesPath, err := testutils.GetTestDataFilePath("netscape-cookies.txt")
	if err != nil {
		t.Fatalf("Failed to load test data file")
	}

	seq, isStrict := TraverseCookies(testCookiesPath)
	cookies, err := seq.ReadAllCookies(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !isStrict() {
		t.Error("file not in strict netscape format")
	}

	if len(cookies) != 3 {
		t.Fatalf("got %d cookies, but expected 3", len(cookies))
	}

	// timezone
	tz := time.Local

	c := cookies[1]
	if c.Domain != ".google.de" {
		t.Errorf("c.Domain=%q", c.Domain)
	}
	if c.Name != "NID" {
		t.Errorf("c.Name=%q", c.Name)
	}
	if c.Path != "/" {
		t.Errorf("c.Path=%q", c.Path)
	}
	if !c.Expires.Equal(time.Date(2021, 4, 16, 15, 33, 4, 0, tz)) {
		t.Errorf("c.Expires=%q", c.Expires)
	}
	if c.Secure {
		t.Error("c.Secure expected false")
	}
	if !c.HttpOnly {
		t.Error("c.HttpOnly expected true")
	}
	// Creation fields unused by parser
	if c.Value != "204=blabla" {
		t.Errorf("c.Value=%q", c.Value)
	}

	c = cookies[2]
	if !c.Secure {
		t.Error("c.Secure expected true")
	}
}
