package firefox

import (
	"context"
	"testing"
	"time"

	"github.com/browserutils/kooky/internal/testutils"
)

func TestReadCookies(t *testing.T) {
	// insert into moz_cookies values
	// (156181,'godoc.org','','GODOC_ORG_SESSION_ID','a748915ba19c6d0b','godoc.org','/github.com/go-sqlite/',1516245891,1516242287597175,1516242287597175,0,0,'');

	testCookiesPath, err := testutils.GetTestDataFilePath("firefox-cookies.sqlite")
	if err != nil {
		t.Fatalf("Failed to load test data file")
	}

	ctx := context.Background()
	cookies, err := TraverseCookies(testCookiesPath).ReadAllCookies(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(cookies) != 1 {
		t.Fatalf("got %d cookies, but expected 1", len(cookies))
	}

	// TZ when I created cookies.sqlite
	tz, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}

	c := cookies[0]
	if c.Domain != "godoc.org" {
		t.Errorf("c.Domain=%q", c.Domain)
	}
	if c.Name != "GODOC_ORG_SESSION_ID" {
		t.Errorf("c.Name=%q", c.Name)
	}
	if c.Path != "/github.com/go-sqlite/" {
		t.Errorf("c.Path=%q", c.Path)
	}
	if !c.Expires.Equal(time.Date(2018, 01, 17, 19, 24, 51, 0, tz)) {
		t.Errorf("c.Expires=%q", c.Expires)
	}
	if c.Secure {
		t.Error("c.Secure expected false")
	}
	if c.HttpOnly {
		t.Error("c.HttpOnly expected false")
	}
	if !c.Creation.Equal(time.Date(2018, 01, 17, 18, 24, 47, 597175000, tz)) {
		t.Errorf("c.Creation=%q", c.Creation)
	}
	if c.Value != "a748915ba19c6d0b" {
		t.Errorf("c.Value=%q", c.Value)
	}

	// 15 columns

	testCookiesPath, err = testutils.GetTestDataFilePath("firefox-v82-linux-cookies.sqlite")
	if err != nil {
		t.Fatalf("Failed to load test data file")
	}

	cookies, err = TraverseCookies(testCookiesPath).ReadAllCookies(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(cookies) != 4 {
		t.Fatalf("got %d cookies, but expected 4", len(cookies))
	}

	tz, err = time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}

	c = cookies[0]
	if c.Domain != "google.de" {
		t.Errorf("c.Domain=%q", c.Domain)
	}
	if c.Name != "NID" {
		t.Errorf("c.Name=%q", c.Name)
	}
	if c.Path != "/" {
		t.Errorf("c.Path=%q", c.Path)
	}
	if !c.Expires.Equal(time.Date(2021, 5, 10, 23, 14, 45, 0, tz)) {
		t.Errorf("c.Expires=%q", c.Expires)
	}
	if !c.Secure {
		t.Error("c.Secure expected true")
	}
	if !c.HttpOnly {
		t.Error("c.HttpOnly expected true")
	}
	if !c.Creation.Equal(time.Date(2020, 11, 8, 22, 14, 45, 143525000, tz)) {
		t.Errorf("c.Creation=%q", c.Creation)
	}
	if c.Value != "204=blabla" {
		t.Errorf("c.Value=%q", c.Value)
	}
}
