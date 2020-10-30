package ie

import (
	"bufio"
	"errors"
	"strconv"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/utils"

	"github.com/bobesa/go-domain-util/domainutil"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s := &ieCookieStore{filename: filename}
	defer s.Close()

	return s.ReadCookies(filters...)
}

func (s *ieCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.open(); err != nil {
		return nil, err
	} else if s.file == nil {
		return nil, errors.New(`file is nil`)
	}

	var (
		lineNr             int
		expLeast, crtLeast uint64
		cookie             *kooky.Cookie
		cookies            []*kooky.Cookie
	)
	scanner := bufio.NewScanner(s.file)
	for scanner.Scan() {
		lineNr = lineNr%9 + 1
		line := scanner.Text()
		switch lineNr {
		case 1:
			cookie = &kooky.Cookie{}
			cookie.Name = line
		case 2:
			cookie.Value = line
		case 3:
			cookie.Domain = domainutil.Domain(line)
		case 4:
			flags, err := strconv.ParseUint(line, 10, 32)
			if err != nil {
				return nil, err
			}
			// TODO: is "Secure" encoded in flags?
			cookie.HttpOnly = flags&(1<<13) != 0
		case 5:
			var err error
			expLeast, err = strconv.ParseUint(line, 10, 32)
			if err != nil {
				return nil, err
			}
		case 6:
			expMost, err := strconv.ParseUint(line, 10, 32)
			if err != nil {
				return nil, err
			}
			cookie.Expires = utils.FromFILETIME(int64(((expMost << 32) ^ expLeast)))
		case 7:
			var err error
			crtLeast, err = strconv.ParseUint(line, 10, 32)
			if err != nil {
				return nil, err
			}
		case 8:
			crtMost, err := strconv.ParseUint(line, 10, 32)
			if err != nil {
				return nil, err
			}
			cookie.Creation = utils.FromFILETIME(int64(((crtMost << 32) ^ crtLeast)))
		case 9:
			// Secure (?)
			if line != `*` {
				return nil, errors.New(`cookie record delimiter not "*"`)
			}
			if kooky.FilterCookie(cookie, filters...) {
				cookies = append(cookies, cookie)
			}
		}
	}

	return cookies, nil
}

/*
https://www.digital-detective.net/random-cookie-filenames/
To mitigate the threat, Internet Explorer 9.0.2 now names the cookie files using a randomly-generated alphanumeric string.

// http://hh.diva-portal.org/smash/get/diva2:635743/FULLTEXT02.pdf p6
With Internet Explorer 10, Microsoft changed the way of storing web related
information. Instead of the old index.dat files, Internet Explorer 10 uses an ESE
database called WebCacheV01.dat...

http://index-of.es/Forensic/Forensic%20Analysis%20of%20Microsoft%20Internet%20Explorer%20Cookie%20Files.pdf
# least and most significant switched

https://sourceforge.net/projects/odessa/files/Galleta/20040505_1/galleta_20040505_1.tar.gz/download
wrong times

https://www.consumingexperience.com/2011/09/internet-explorer-cookie-contents-new.html

https://en.wikipedia.org/wiki/Index.dat
does not delete the hidden index.dat file in the Temporary Internet Files directory, which contains a copy of the cookies that were in the Cookies directory

https://tzworks.net/prototype_page.php?proto_id=6
*/
