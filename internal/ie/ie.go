package ie

import (
	"errors"

	"github.com/zellyn/kooky"
)

func (s *CookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	return s.CookieStore.ReadCookies(filters...)
}

/*
NOTES:

Internet Explorer 9 uses index.dat
Internet Explorer 10, 11 use WebCacheV01.dat

https://www.digital-detective.net/random-cookie-filenames/
To mitigate the threat, Internet Explorer 9.0.2 now names the cookie files using a randomly-generated alphanumeric string.

http://index-of.es/Forensic/Forensic%20Analysis%20of%20Microsoft%20Internet%20Explorer%20Cookie%20Files.pdf
# least and most significant switched

https://sourceforge.net/projects/odessa/files/Galleta/20040505_1/galleta_20040505_1.tar.gz/download
wrong times

https://www.consumingexperience.com/2011/09/internet-explorer-cookie-contents-new.html

# test VMs
https://web.archive.org/web/20150410044551/https://www.modern.ie/en-us
https://developer.microsoft.com/en-us/microsoft-edge/tools/vms/
*/
