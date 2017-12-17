# kooky

Package kooky contains routines to reach into cookie stores for Chrome
and Safari, and retrieve the cookies.

You don't usually want to do this, but when you do, you do. Apparently
it's common enough that there is example code scattered around the net
in various languages. If you wanted it in Go, you're in luck.

It aspires to be pure Go (I spent quite a while making
[go-sqlite/sqlite3](https://github.com/go-sqlite/sqlite3) work for
it), but I guess the keychain parts
([keybase/go-keychain](http://github.com/keybase/go-keychain)) mess that up.

## Status

[![No Maintenance Intended](http://unmaintained.tech/badge.svg)](http://unmaintained.tech/)

Basic functionality works, on MacOS. I expect Linux to work too, since
it doesn't encrypt.

PRs more than welcome.

### TODOs

- [ ] Make it work on Windows. (Look at
      [this](https://play.golang.org/p/fknP9AuLU-) and
      [this](https://github.com/cfstras/chromecsv/blob/master/crypt_windows.go)
      to learn how to decrypt.)
- [ ] Handle rows in Chrome's cookie DB with other than 14 columns (?)
- [ ] Make it work for Firefox.
