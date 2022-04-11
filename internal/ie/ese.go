package ie

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/timex"

	"github.com/Velocidex/ordereddict"
	"www.velocidex.com/golang/go-ese/parser"
)

func (s *ESECookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.Open(); err != nil {
		return nil, err
	} else if s.File == nil {
		return nil, errors.New(`file is nil`)
	}

	if s.ESECatalog == nil {
		return nil, errors.New(`ESE catalog is nil`)
	}

	tables := s.ESECatalog.Tables
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
		if err := s.ESECatalog.DumpTable(tableName, cbCookieEntries); err != nil {
			errs = append(errs, err)
			err = errorList{Errors: errs}
			return nil, err
		}
	}

	return cookies, nil
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
	// createdIn    := entry.flags&(1<<6)  != 0 // server: false, client true
	// unkownFlag01 := entry.flags&(1<<10) != 0
	cookie.HttpOnly = entry.flags&(1<<13) != 0
	// hostOnly     := entry.flags&(1<<14) != 0 // TODO: is this SameSite?
	// unkownFlag02 := entry.flags&(1<<19) != 0
	// unkownFlag03 := entry.flags&(1<<31) != 0

	cookie.Expires = timex.FromFILETIME(entry.expires)

	// TODO: use "CookieEntryEx_##.LastModified" field as "Cookie.Creation" time?

	return cookie, nil
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

// TODO: create temporary copy of the ESE file on Windows - a service on Windows has a permanent lock on it, remove temporary copy in Close()

/*
	// TODO: cookie text file locations stored in WebCacheV01.dat ?
	//
	// directories with old text file cookies
	cookiesContainers, _ := getEdgeCookieDirectories(edgeESECookieStore.ESECatalog)
	for _, cookiesContainer := range cookiesContainers {
		_ = cookiesContainer.directory // TODO
	}
*/
