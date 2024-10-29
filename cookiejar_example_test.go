package kooky_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/xiazemin/kooky"
	_ "github.com/xiazemin/kooky/browser/firefox"
)

func Example_cookieJar() {
	ctx := context.TODO()
	stores := kooky.FindAllCookieStores(ctx)
	var s kooky.CookieStore
	for _, store := range stores {
		if store.Browser() != `firefox` || !store.IsDefaultProfile() {
			continue
		}
		s = store
		break
	}
	// jar := s
	// only store cookies relevant for the target website in the cookie jar
	jar, _ := s.SubJar(ctx, kooky.FilterFunc(func(c *kooky.Cookie) bool {
		return kooky.Domain(`github.com`).Filter(c) || kooky.Domain(`.github.com`).Filter(c)
	}))

	u, _ := url.Parse(`https://github.com/settings/profile`)

	cookies := kooky.FilterCookies(ctx, jar.Cookies(u), kooky.Name(`logged_in`)).Collect(ctx)
	if len(cookies) == 0 {
		log.Fatal(`not logged in`)
	}

	client := http.Client{Jar: jar}
	resp, _ := client.Get(u.String())
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), `id="user_profile_name"`) {
		fmt.Print("not ")
	}
	fmt.Println("logged in")

	// Output: logged in
}
