package edge

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/utils"

	"github.com/Velocidex/ordereddict"
	"www.velocidex.com/golang/go-ese/parser"
)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s := &edgeCookieStore{filename: filename}
	defer s.Close()

	return s.ReadCookies(filters...)
}

func (s *edgeCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.open(); err != nil {
		return nil, err
	} else if s.file == nil {
		return nil, errors.New(`file is nil`)
	}

	// In the file header of the database we find that the first 4 bytes are a XOR checksum.
	// The following 4 bytes after the checksum is a file signature. The file signature has
	// offset 4, and the value is EF CD AB 89.
	signature := make([]byte, 8)
	if _, err := s.file.Read(signature); err != nil {
		return nil, err
	}
	if _, err := s.file.Seek(0, 0); err != nil {
		return nil, err
	}
	var signatureESEdatabase = []byte{239, 205, 171, 137} // EF CD AB 89
	if !bytes.Equal(signature[4:8], signatureESEdatabase) {
		return nil, errors.New(`file is not an ESE database`)
	}

	ese_ctx, err := parser.NewESEContext(s.file)
	if err != nil {
		return nil, err
	}
	catalog, err := parser.ReadCatalog(ese_ctx)
	if err != nil {
		return nil, err
	}

	// directories with old text file cookies
	cookiesContainers, _ := getEdgeCookieDirectories(catalog)
	for _, cookiesContainer := range cookiesContainers {
		_ = cookiesContainer.directory // TODO
	}

	var cookies []*kooky.Cookie
	textCookies, errTXT := getEdgeTextCookies(catalog, filters...)
	cookies = append(cookies, textCookies...)

	eseCookies, errESE := getEdgeESEcookies(catalog, filters...)
	cookies = append(cookies, eseCookies...)

	chromeCookies, errCHR := getEdgeChromecookies(s.filename, filters...)
	cookies = append(cookies, chromeCookies...)

	if errTXT != nil && errESE != nil && errCHR != nil && len(cookies) == 0 {
		return nil, errors.New(`cannot read edge cookies file`)
	}

	return cookies, nil
}

type webCacheContainer struct {
	containerID      int    // ContainerId
	setID            int    // SetId
	flags            int    // Flags
	size             int    // Size
	limit            int    // Limit
	lastScavengeTime int    // LastScavengeTime
	entryMaxAge      int    // EntryMaxAge
	lastAccessTime   int64  // LastAccessTime
	name             string // Name
	partitionID      string // PartitionId
	directory        string // Directory
}

func getEdgeCookieDirectories(catalog *parser.Catalog) ([]webCacheContainer, error) {
	var cookiesContainers []webCacheContainer

	cbContainers := func(row *ordereddict.Dict) error {
		var name, directory string
		if n, ok := row.GetString(`Name`); ok {
			name = strings.TrimRight(parser.UTF16BytesToUTF8([]byte(n), binary.LittleEndian), "\x00")
		} else {
			return nil
		}
		if name != `Cookies` {
			return nil
		}

		directory, ok := row.GetString(`Directory`)
		if !ok {
			return nil
		}

		cookiesContainers = append(
			cookiesContainers,
			webCacheContainer{
				name:      name,
				directory: directory,
			},
		)

		return nil
	}

	if err := catalog.DumpTable(`Containers`, cbContainers); err != nil {
		return nil, err
	}

	return cookiesContainers, nil
}

type webCacheCookieEntry struct {
	entryID                uint64 // EntryId
	minimizedRDomainLength uint64 // MinimizedRDomainLength
	flags                  uint32 // Flags
	expires                int64  // Expires
	lastModified           int64  // LastModified
	cookieHash             string // CookieHash
	rDomain                string // RDomain
	path                   string // Path
	name                   string // Name
	value                  string // Value
}

func getEdgeESEcookies(catalog *parser.Catalog, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	tables := catalog.Tables
	if tables == nil {
		return nil, errors.New(`catalog.Tables is nil`)
	}

	var cookies []*kooky.Cookie
	var errs []error

	cbCookieEntries := func(row *ordereddict.Dict) error {
		if row == nil {
			errs = append(errs, errors.New(`row is nil`))
			return nil
		}

		var cookieEntry webCacheCookieEntry
		if e, ok := row.GetInt64(`EntryId`); ok {
			cookieEntry.entryID = uint64(e)
		} else {
			errs = append(errs, errors.New(`no int64 EntryId`))
			return nil
		}
		if e, ok := row.GetInt64(`MinimizedRDomainLength`); ok {
			cookieEntry.minimizedRDomainLength = uint64(e)
		} else {
			errs = append(errs, errors.New(`no int64 MinimizedRDomainLength`))
			return nil
		}
		if e, ok := row.GetInt64(`Flags`); ok {
			cookieEntry.flags = uint32(e)
		} else {
			errs = append(errs, errors.New(`no int64 Flags`))
			return nil
		}
		if e, ok := row.GetInt64(`Expires`); ok {
			cookieEntry.expires = e
		} else {
			errs = append(errs, errors.New(`no int64 Expires`))
			return nil
		}
		if e, ok := row.GetInt64(`LastModified`); ok {
			cookieEntry.lastModified = e
		} else {
			errs = append(errs, errors.New(`no int64 LastModified`))
			return nil
		}
		var ok bool
		cookieEntry.cookieHash, ok = row.GetString(`CookieHash`)
		if !ok {
			errs = append(errs, errors.New(`no string CookieHash`))
			return nil
		}
		cookieEntry.rDomain, ok = row.GetString(`RDomain`)
		if !ok {
			errs = append(errs, errors.New(`no string RDomain`))
			return nil
		}
		cookieEntry.path, ok = row.GetString(`Path`)
		if !ok {
			errs = append(errs, errors.New(`no string Path`))
			return nil
		}
		cookieEntry.name, ok = row.GetString(`Name`)
		if !ok {
			errs = append(errs, errors.New(`no string Name`))
			return nil
		}
		cookieEntry.value, ok = row.GetString(`Value`)
		if !ok {
			errs = append(errs, errors.New(`no string Value`))
			return nil
		}

		cookie, err := convertCookieEntry(&cookieEntry)
		if err != nil {
			errs = append(errs, err)
			return nil
		}

		if kooky.FilterCookie(cookie, filters...) {
			cookies = append(cookies, cookie)
		}

		return nil
	}

	for _, tableName := range tables.Keys() {
		if !strings.HasPrefix(tableName, `CookieEntryEx_`) {
			continue
		}
		if err := catalog.DumpTable(tableName, cbCookieEntries); err != nil {
			errs = append(errs, err)
			err = errorList{Errors: errs}
			return nil, err
		}
	}

	return cookies, nil
}

type errorList struct {
	Errors []error
}

func (l errorList) Error() string {
	if len(l.Errors) > 0 {
		return l.Errors[0].Error() + `, additional errors...`
	}
	return `unknown error`
}

func eseHexDecodeString(raw string) (string, error) {
	b, err := hex.DecodeString(raw)
	if err != nil {
		return ``, err
	}
	return strings.Split(string(b), "\x00")[0], nil
}

func convertCookieEntry(entry *webCacheCookieEntry) (*kooky.Cookie, error) {
	if entry == nil {
		return nil, errors.New(`cookie entry is nil`)
	}

	cookie := &kooky.Cookie{}
	var err error
	cookie.Name, err = eseHexDecodeString(entry.name)
	if err != nil {
		return nil, err
	}

	cookie.Value, err = eseHexDecodeString(entry.value)
	if err != nil {
		return nil, err
	}

	cookie.Path = entry.path

	rdp := strings.Split(strings.Trim(entry.rDomain, `.`), `.`)
	for i := 0; i < len(rdp); i++ {
		cookie.Domain += `.` + rdp[len(rdp)-1-i]
	}
	cookie.Domain = strings.TrimLeft(cookie.Domain, `.`)

	cookie.Secure = entry.flags&1 != 0
	// createdIn    := entry.flags&(1<<6)  != 0 // TODO: server: false, client true
	// unkownFlag01 := entry.flags&(1<<10) != 0 // TODO
	cookie.HttpOnly = entry.flags&(1<<13) != 0
	// hostOnly     := entry.flags&(1<<14) != 0 // TODO: is this SameSite?
	// unkownFlag02 := entry.flags&(1<<19) != 0 // TODO
	// unkownFlag03 := entry.flags&(1<<31) != 0 // TODO

	cookie.Expires = utils.FromFILETIME(entry.expires)

	// TODO: use "CookieEntryEx_##.LastModified" field as "Cookie.Creation" time?

	return cookie, nil
}

func getEdgeTextCookies(catalog *parser.Catalog, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	return nil, errors.New(`not implemented`)
}

func getEdgeChromecookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	return nil, errors.New(`not implemented`)
}

/*
tools used for dev:
https://www.nirsoft.net/utils/edge_cookies_view.html
https://www.nirsoft.net/utils/ese_database_view.html
*/
