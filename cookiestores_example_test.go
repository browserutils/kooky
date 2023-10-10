package kooky_test

import (
	"fmt"

	"github.com/xiazemin/kooky"
	_ "github.com/xiazemin/kooky/browser/all" // register cookiestore finders
)

func ExampleFindAllCookieStores() {
	cookieStores := kooky.FindAllCookieStores()

	for _, store := range cookieStores {
		// CookieStore keeps files/databases open for repeated reads
		// close those when no longer needed
		defer store.Close()

		var filters = []kooky.Filter{
			kooky.Valid, // remove expired cookies
		}

		cookies, _ := store.ReadCookies(filters...)
		for _, cookie := range cookies {
			fmt.Printf(
				"%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				store.Browser(),
				store.Profile(),
				store.FilePath(),
				cookie.Domain,
				cookie.Name,
				cookie.Value,
				cookie.Expires.Format(`2006.01.02 15:04:05`),
			)
		}
	}
}
