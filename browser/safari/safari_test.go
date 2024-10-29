package safari

import (
	"context"
	"testing"
	"time"

	"github.com/xiazemin/kooky"
	"github.com/xiazemin/kooky/internal/testutils"
)

// d18f6247db68045dfbab126d814baf2cf1512141391
func TestReadCookies(t *testing.T) {
	testCookiesPath, err := testutils.GetTestDataFilePath("safari-macos-cookie-db.binarycookies")
	if err != nil {
		t.Fatalf("Failed to load test data file")
	}

	ctx := context.Background()
	cookies, err := TraverseCookies(testCookiesPath).ReadAllCookies(ctx)
	if err != nil {
		t.Fatal(err)
	}

	domain := "news.ycombinator.com"
	name := "user"
	cookies = kooky.FilterCookies(ctx, cookies, kooky.Domain(domain), kooky.Name(name)).Collect(ctx)
	if len(cookies) == 0 {
		t.Fatalf("Found no cookies with domain=%q, name=%q", domain, name)
	}
	cookie := cookies[0]
	wantValue := "zellyn&EdK9mzRM38fGtIZQTiqCyAeWg93RDjdo"
	if cookie.Value != wantValue {
		t.Errorf("Want cookie value %q; got %q", wantValue, cookie.Value)
	}

	wantExpires := time.Date(2038, 01, 17, 19, 14, 07, 0, time.UTC)
	if !cookie.Expires.Equal(wantExpires) {
		t.Errorf("Want cookie.Expires=%v; got %v", wantExpires, cookie.Expires)
	}

	wantCreation := time.Date(2017, 12, 16, 23, 23, 19, 0, time.UTC)
	if !cookie.Creation.Equal(wantCreation) {
		t.Errorf("Want cookie.Creation=%v; got %v", wantCreation, cookie.Creation)
	}
}
