# kooky

[![PkgGoDev](https://pkg.go.dev/badge/github.com/browserutils/kooky)](https://pkg.go.dev/github.com/browserutils/kooky)
[![Go Report Card](https://goreportcard.com/badge/browserutils/kooky)](https://goreportcard.com/report/browserutils/kooky)
![Lines of code](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fapi.codetabs.com%2Fv1%2Floc%2F%3Fgithub%3Dbrowserutils%2Fkooky%26ignored%3Dvendor%2Ctestdata&query=%24%5B%3F(%40.language%3D%3D%22Go%22)%5D.linesOfCode&logo=Go&label=lines%20of%20code&cacheSeconds=3600)
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

## Example usage

### Any Browser - Cookie Filter Usage

```go
package main

import (
	"fmt"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all" // register cookie store finders!
)

func main() {
	// uses registered finders to find cookie store files in default locations
	// applies the passed filters "Valid", "DomainHasSuffix()" and "Name()" in order to the cookies
	cookiesSeq := kooky.TraverseCookies(context.TODO(), kooky.Valid, kooky.DomainHasSuffix(`google.com`), kooky.Name(`NID`)).OnlyCookies()

	for cookie := range cookiesSeq {
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

	"github.com/browserutils/kooky/browser/chrome"
)

func main() {
	dir, _ := os.UserConfigDir() // "/<USER>/Library/Application Support/"
	cookiesFile := dir + "/Google/Chrome/Default/Cookies"
	cookiesSeq := chrome.TraverseCookies(cookiesFile).OnlyCookies()
	for cookie := range cookiesSeq {
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

	"github.com/browserutils/kooky/browser/safari"
)

func main() {
	dir, _ := os.UserHomeDir()
	cookiesFile := dir + "/Library/Containers/com.apple.Safari/Data/Library/Cookies/Cookies.binarycookies"
	cookiesSeq := safari.TraverseCookies(cookiesFile).OnlyCookies()
	for cookie := range cookiesSeq {
		fmt.Println(cookie)
	}
}
```

## Thanks/references
- Thanks to [@dacort](https://github.com/dacort) for MacOS cookie decrypting
  code at https://gist.github.com/dacort/bd6a5116224c594b14db.
- Thanks to [@as0ler](https://github.com/as0ler)
  (and originally [@satishb3](https://github.com/satishb3) I believe) for
  Safari cookie-reading Python code at https://github.com/as0ler/BinaryCookieReader.
- Thanks to all the people who have contributed functionality and fixes:
  - [@srlehn](https://github.com/srlehn) - many fixes, Linux support for Chrome, added about a dozen browsers!
  - [@zippoxer](https://github.com/zippoxer) - Windows support for Chrome
  - [@adamdecaf](https://github.com/adamdecaf) - Firefox support
  - [@barnardb](https://github.com/barnardb) - better row abstraction, fixing column length errors
