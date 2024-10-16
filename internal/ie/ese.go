package ie

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/iterx"
	"github.com/browserutils/kooky/internal/timex"

	"github.com/Velocidex/ordereddict"
	"www.velocidex.com/golang/go-ese/parser"
)

func (s *ESECookieStore) TraverseCookies(filters ...kooky.Filter) kooky.CookieSeq {
	if s == nil {
		return iterx.ErrCookieSeq(errors.New(`cookie store is nil`))
	}
	if err := s.Open(); err != nil {
		return iterx.ErrCookieSeq(err)
	} else if s.File == nil {
		return iterx.ErrCookieSeq(errors.New(`file is nil`))
	}

	if s.ESECatalog == nil {
		return iterx.ErrCookieSeq(errors.New(`ESE catalog is nil`))
	}

	tables := s.ESECatalog.Tables
	if tables == nil {
		return iterx.ErrCookieSeq(errors.New(`catalog.Tables is nil`))
	}

	cbCookieEntries := func(yield func(*kooky.Cookie, error) bool) func(row *ordereddict.Dict) error {
		return func(row *ordereddict.Dict) error {
			if row == nil {
				if !yield(nil, errors.New(`row is nil`)) {
					return iterx.ErrYieldEnd
				}
				return nil
			}

			var cookieEntry webCacheCookieEntry
			if e, ok := row.GetInt64(`EntryId`); ok {
				cookieEntry.entryID = uint64(e)
			} else {
				if !yield(nil, errors.New(`no int64 EntryId`)) {
					return iterx.ErrYieldEnd
				}
				return nil
			}
			if e, ok := row.GetInt64(`MinimizedRDomainLength`); ok {
				cookieEntry.minimizedRDomainLength = uint64(e)
			} else {
				if !yield(nil, errors.New(`no int64 MinimizedRDomainLength`)) {
					return iterx.ErrYieldEnd
				}
				return nil
			}
			if e, ok := row.GetInt64(`Flags`); ok {
				cookieEntry.flags = uint32(e)
			} else {
				if !yield(nil, errors.New(`no int64 Flags`)) {
					return iterx.ErrYieldEnd
				}
				return nil
			}
			if e, ok := row.GetInt64(`Expires`); ok {
				cookieEntry.expires = e
			} else {
				if !yield(nil, errors.New(`no int64 Expires`)) {
					return iterx.ErrYieldEnd
				}
				return nil
			}
			if e, ok := row.GetInt64(`LastModified`); ok {
				cookieEntry.lastModified = e
			} else {
				if !yield(nil, errors.New(`no int64 LastModified`)) {
					return iterx.ErrYieldEnd
				}
				return nil
			}
			var ok bool
			cookieEntry.cookieHash, ok = row.GetString(`CookieHash`)
			if !ok {
				if !yield(nil, errors.New(`no string CookieHash`)) {
					return iterx.ErrYieldEnd
				}
				return nil
			}
			cookieEntry.rDomain, ok = row.GetString(`RDomain`)
			if !ok {
				if !yield(nil, errors.New(`no string RDomain`)) {
					return iterx.ErrYieldEnd
				}
				return nil
			}
			cookieEntry.path, ok = row.GetString(`Path`)
			if !ok {
				if !yield(nil, errors.New(`no string Path`)) {
					return iterx.ErrYieldEnd
				}
				return nil
			}
			cookieEntry.name, ok = row.GetString(`Name`)
			if !ok {
				if !yield(nil, errors.New(`no string Name`)) {
					return iterx.ErrYieldEnd
				}
				return nil
			}
			cookieEntry.value, ok = row.GetString(`Value`)
			if !ok {
				if !yield(nil, errors.New(`no string Value`)) {
					return iterx.ErrYieldEnd
				}
				return nil
			}

			cookie, errCookie := convertCookieEntry(&cookieEntry, s)
			if !iterx.CookieFilterYield(context.Background(), cookie, errCookie, yield, filters...) {
				return nil
			}

			return nil
		}
	}

	seq := func(yield func(*kooky.Cookie, error) bool) {
		for _, tableName := range tables.Keys() {
			if !strings.HasPrefix(tableName, `CookieEntryEx_`) {
				continue
			}
			if err := s.ESECatalog.DumpTable(tableName, cbCookieEntries(yield)); err != nil {
				if !yield(nil, err) {
					return
				}
			}
		}
	}
	return seq
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

var _ error = (*errorList)(nil)

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

func convertCookieEntry(entry *webCacheCookieEntry, bi kooky.BrowserInfo) (*kooky.Cookie, error) {
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

	cookie.Browser = bi

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
