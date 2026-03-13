package kooky_test

import (
	"context"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all" // register cookiestore finders
	"github.com/browserutils/kooky/filter"
)

var cookieName = `NID`

func ExampleFilterCookies() {
	ctx := context.TODO()

	cookies := kooky.AllCookies(). // automatic read
					Seq().
					Filter(
			ctx,
			filter.Valid,                    // remove expired cookies
			filter.DomainContains(`google`), // cookie domain has to contain "google"
			filter.Name(cookieName),         // cookie name is "NID"
			filter.Debug,                    // print cookies after applying previous filter
		).
		Collect(ctx) // iterate and collect in a slice

	_ = cookies // do something
}
