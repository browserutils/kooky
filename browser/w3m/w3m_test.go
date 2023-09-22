package w3m

import (
	"testing"
	"time"

	"github.com/browserutils/kooky/internal/testutils"
)

func TestReadCookies(t *testing.T) {
	testCookiesPath, err := testutils.GetTestDataFilePath("w3m-cookie")
	if err != nil {
		t.Fatalf("Failed to load test data file")
	}

	cookies, err := ReadCookies(testCookiesPath)
	if err != nil {
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
	if c.Name != "CGIC" {
		t.Errorf("c.Name=%q", c.Name)
	}
	if c.Path != "/complete/search" {
		t.Errorf("c.Path=%q", c.Path)
	}
	if !c.Expires.Equal(time.Date(2021, 4, 13, 10, 13, 04, 0, tz)) {
		t.Errorf("c.Expires=%q", c.Expires)
	}
	if c.Secure {
		t.Error("c.Secure expected false")
	}
	// HttpOnly, Creation fields unused by parser
	if c.Value != "blabla" {
		t.Errorf("c.Value=%q", c.Value)
	}
}
