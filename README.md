# kooky

[![PkgGoDev](https://pkg.go.dev/badge/github.com/zellyn/kooky)](https://pkg.go.dev/github.com/zellyn/kooky)
[![Go Report Card](https://goreportcard.com/badge/zellyn/kooky)](https://goreportcard.com/report/zellyn/kooky)
![Lines of code](https://img.shields.io/tokei/lines/github/zellyn/kooky)
[![No Maintenance Intended](http://unmaintained.tech/badge.svg)](http://unmaintained.tech/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)
[![MIT license](https://img.shields.io/badge/License-MIT-blue.svg)](https://lbesson.mit-license.org/)

Reaching into browser-specific, vaguely documented, possibly
concurrently modified cookie stores to pilfer cookies is a bad idea.
Since you've arrived here, you're almost certainly going to do it
anyway. Me too. And if we're going to do the Wrong Thing, at least
let's try to Do it Right.

Package kooky contains routines to reach into cookie stores for Chrome, Firefox, Safari, ... and retrieve the cookies.

It aspires to be pure Go (I spent quite a while making
[go-sqlite/sqlite3](https://github.com/go-sqlite/sqlite3) work for
it).

It also aspires to work for all major browsers, on all three
major platforms.

## Status

Basic functionality works on Windows, MacOS and Linux.
Some functions might not yet be implemented on some platforms.
**The API is currently not expected to be at all stable.**

PRs more than welcome.

## TODOs

- [ ] Set up CI
- [x] Make it work on Windows. (Look at
      [this](https://play.golang.org/p/fknP9AuLU-) and
      [this](https://github.com/cfstras/chromecsv/blob/master/crypt_windows.go)
      to learn how to decrypt.)
- [x] Handle rows in Chrome's cookie DB with other than 14 columns (?)

## Example usage

### Any Browser - Cookie Filter Usage

```go
package main

import (
	"fmt"

	"github.com/zellyn/kooky"
	_ "github.com/zellyn/kooky/browser/all" // register cookie store finders!
)

func main() {
	// uses registered finders to find cookie store files in default locations
	// applies the passed filters "Valid", "DomainHasSuffix()" and "Name()" in order to the cookies
	cookies := kooky.ReadCookies(kooky.Valid, kooky.DomainHasSuffix(`google.com`), kooky.Name(`NID`))

	for _, cookie := range cookies {
		fmt.Println(cookie.Domain, cookie.Name, cookie.Value)
	}
 }
```

### Chrome on macOS

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/zellyn/kooky/browser/chrome"
)

func main() {
	dir, _ := os.UserConfigDir() // "/<USER>/Library/Application Support/"
	cookiesFile := dir + "/Google/Chrome/Default/Cookies"
	cookies, err := chrome.ReadCookies(cookiesFile)
	if err != nil {
		log.Fatal(err)
	}
	for _, cookie := range cookies {
		fmt.Println(cookie)
	}
}
```

### Safari

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/zellyn/kooky/browser/safari"
)

func main() {
	dir, _ := os.UserHomeDir()
	cookiesFile := dir + "/Library/Cookies/Cookies.binarycookies"
	cookies, err := safari.ReadCookies(cookiesFile)
	if err != nil {
		log.Fatal(err)
	}
	for _, cookie := range cookies {
		fmt.Println(cookie)
	}
}
```

## Thanks/references
- Thanks to [@dacort](http://github.com/dacort) for MacOS cookie decrypting
  code at https://gist.github.com/dacort/bd6a5116224c594b14db.
- Thanks to [@as0ler](http://github.com/as0ler)
  (and originally [@satishb3](http://github.com/satishb3) I believe) for
  Safari cookie-reading Python code at https://github.com/as0ler/BinaryCookieReader.
- Thanks to all the people who have contributed functionality and fixes:
  - [@srlehn](http://github.com/srlehn) - many fixes, Linux support for Chrome, added about a dozen browsers!
  - [@zippoxer](http://github.com/zippoxer) - Windows support for Chrome
  - [@adamdecaf](http://github.com/adamdecaf) - Firefox support
  - [@barnardb](https://github.com/barnardb) - better row abstraction, fixing column length errors
