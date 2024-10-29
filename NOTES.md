### TODO

- [ ] Set up CI
- [x] Make it work on Windows. (Look at
      [this](https://play.golang.org/p/fknP9AuLU-) and
      [this](https://github.com/cfstras/chromecsv/blob/master/crypt_windows.go)
      to learn how to decrypt.)
- [x] Handle rows in Chrome's cookie DB with other than 14 columns (?)
- [ ] Figure out what to do with quoted values, like the "bcookie" cookie from slideshare.net
  (related (?): [go issue #46443:  net/http: add field Cookie.Quoted bool](https://github.com/golang/go/issues/46443))
