package kooky_test

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/zellyn/kooky"
	_ "github.com/zellyn/kooky/browser/firefox"
)

func Example_cookieJar() {
	stores := kooky.FindAllCookieStores()
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
	jar, _ := s.SubJar(kooky.Domain(`github.com`))

	u, _ := url.Parse(`https://github.com/settings/profile`)
	var loggedIn bool
	cookies := kooky.FilterCookies(jar.Cookies(u), kooky.Name(`logged_in`))
	if len(cookies) > 0 {
		loggedIn = true
	}
	if !loggedIn {
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
