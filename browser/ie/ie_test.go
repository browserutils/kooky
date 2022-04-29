package ie

import (
	"testing"
	"time"

	"github.com/zellyn/kooky/internal/testutils"
)

func TestReadCookies(t *testing.T) {
	testCookiesPath, err := testutils.GetTestDataFilePath("ie-user@google[4].txt")
	if err != nil {
		t.Fatalf("Failed to load test data file")
	}

	cookies, err := ReadCookies(testCookiesPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(cookies) != 3 {
		t.Fatalf("got %d cookies, but expected 3", len(cookies))
	}

	// timezone
	tz, _ := time.LoadLocation("Europe/Berlin")

	c := cookies[0]
	if c.HttpOnly {
		t.Error("c.HttpOnly expected false")
	}

	c = cookies[1]
	if c.Domain != "google.de" {
		t.Errorf("c.Domain=%q", c.Domain)
	}
	if c.Name != "NID" {
		t.Errorf("c.Name=%q", c.Name)
	}
	if c.Path != "/" {
		t.Errorf("c.Path=%q", c.Path)
	}
	if !c.Expires.Equal(time.Date(2021, 3, 2, 18, 56, 24, 0, tz)) {
		t.Errorf("c.Expires=%q", c.Expires)
	}
	if !c.Creation.Equal(time.Date(2020, 8, 31, 19, 56, 24, 256e6, tz)) {
		t.Errorf("c.Creation=%q", c.Creation)
	}
	if !c.HttpOnly {
		t.Error("c.HttpOnly expected true")
	}
	// Secure field unused by parser
	if c.Value != "204=blabla" {
		t.Errorf("c.Value=%q", c.Value)
	}

	c = cookies[2]
	if c.Path != "/complete/search" {
		t.Errorf("c.Path=%q", c.Path)
	}
}
