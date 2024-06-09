package uzbl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/browserutils/kooky/browser/netscape"
	"github.com/browserutils/kooky/internal/testutils"
)

func TestReadCookies(t *testing.T) {
	testCookiesPath, err := testutils.GetTestDataFilePath("uzbl-cookies.txt")
	if err != nil {
		t.Fatalf("Failed to load test data file")
	}
	cookies, err := TraverseCookies(testCookiesPath).ReadAllCookies(context.Background())
	if err != nil && !errors.Is(err, netscape.ErrNotStrict) {
		t.Fatal(err)
	}
	if len(cookies) != 2 {
		t.Fatalf("got %d cookies, but expected 2", len(cookies))
	}

	// timezone
	tz := time.Local

	c := cookies[0]
	if c.Domain != ".google.de" {
		t.Errorf("c.Domain=%q", c.Domain)
	}
	if c.Name != "NID" {
		t.Errorf("c.Name=%q", c.Name)
	}
	if c.Path != "/" {
		t.Errorf("c.Path=%q", c.Path)
	}
	if !c.Expires.Equal(time.Date(2021, 4, 17, 18, 22, 55, 0, tz)) {
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

	testCookiesPath, err = testutils.GetTestDataFilePath("uzbl-session-cookies.txt")
	if err != nil {
		t.Fatalf("Failed to load test data file")
	}
	cookies, err = TraverseCookies(testCookiesPath).ReadAllCookies(context.Background())
	if err != nil && !errors.Is(err, netscape.ErrNotStrict) {
		t.Fatal(err)
	}
	if len(cookies) != 1 {
		t.Fatalf("got %d cookies, but expected 1", len(cookies))
	}

	c = cookies[0]
	if c.Domain != ".youtube.com" {
		t.Errorf("c.Domain=%q", c.Domain)
	}
	if c.Name != "YSC" {
		t.Errorf("c.Name=%q", c.Name)
	}
	if c.Path != "/" {
		t.Errorf("c.Path=%q", c.Path)
	}
	// Expires field empty in "session-cookies.txt" file
	if !c.Secure {
		t.Error("c.Secure expected true")
	}
	if !c.HttpOnly {
		t.Error("c.HttpOnly expected true")
	}
	if c.Value != "blabla" {
		t.Errorf("c.Value=%q", c.Value)
	}
}
