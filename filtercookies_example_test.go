package kooky_test

import (
	"context"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all" // register cookiestore finders
)

var cookieName = `NID`

func ExampleFilterCookies() {
	ctx := context.TODO()

	cookies := kooky.AllCookies(). // automatic read
					Seq().
					Filter(
			ctx,
			kooky.Valid,                    // remove expired cookies
			kooky.DomainContains(`google`), // cookie domain has to contain "google"
			kooky.Name(cookieName),         // cookie name is "NID"
			kooky.Debug,                    // print cookies after applying previous filter
		).
		Collect(ctx) // iterate and collect in a slice

	_ = cookies // do something
}
