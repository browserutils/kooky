package ie

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/bytesx"
	"github.com/zellyn/kooky/internal/cookies"
	"github.com/zellyn/kooky/internal/timex"
)

// index.dat parser

func (s *IECacheCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.Open(); err != nil {
		return nil, err
	} else if s.File == nil {
		return nil, errors.New(`file is nil`)
	}

	ieCacheVersion, err := bytesx.ReadString(s.File, "file format version", 0x00, 0x18)
	if err != nil {
		return nil, err
	}
	if ieCacheVersion != `5.2` {
		return nil, errors.New(`unsupported IE url cache version`)
	}

	offsetHashStart, err := bytesx.ReadOffSetInt64LE(s.File, "hash table offset", 0x00, 0x20)
	if err != nil {
		return nil, err
	}
	hashSig, err := bytesx.ReadBytesN(s.File, "hash table offset", offsetHashStart, 0x00, 4)
	if err != nil {
		return nil, err
	}
	if string(hashSig) != `HASH` {
		return nil, errors.New(`wrong offset for hash table`)
	}
	// TODO: use hash entries for domain search

	var entries []*CacheCookieEntry
	var textCookieStores []*TextCookieStore
	// TODO: do url records always start at 0x5000?
	urlSig := []byte(`URL `)
	s.File.Seek(0, io.SeekStart)
	for {
		offsetURLEntry, err := scanRest(s.File, urlSig)
		if err != nil {
			break
		}
		entry := &CacheCookieEntry{}

		blockCount, err := bytesx.ReadOffSetInt64LE(s.File, "block count", offsetURLEntry, 4)
		if err != nil {
			return nil, err
		}
		dateModification, err := getFILETIME(s.File, "modification date", offsetURLEntry, 8)
		if err != nil {
			return nil, err
		}
		dateLastAccessed, err := getFILETIME(s.File, "last access date", offsetURLEntry, 16)
		if err != nil {
			return nil, err
		}
		// probably less accurate copy of DateLastAccessed
		dateLastChecked, err := getFATTIME(s.File, "last check date", offsetURLEntry, 80)
		if err != nil {
			return nil, err
		}
		dateExpiry, err := getFATTIME(s.File, "expiry date", offsetURLEntry, 24)
		if err != nil {
			return nil, err
		}

		entry.BlockCount = blockCount
		entry.DateModification = dateModification
		entry.DateLastAccessed = dateLastAccessed
		entry.DateLastChecked = dateLastChecked
		entry.DateExpiry = dateExpiry

		// https://github.com/libyal/libmsiecf/blob/main/documentation/MSIE%20Cache%20File%20(index.dat)%20format.asciidoc#44-url-record-types
		// Cookie:<username>@<URI>
		offsetURLRecordLocation, err := bytesx.ReadOffSetInt64LE(s.File, "location offset", offsetURLEntry, 52) // always 104?
		if err != nil {
			return nil, err
		}
		location, err := bytesx.ReadString(s.File, "location", offsetURLEntry, offsetURLRecordLocation)
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(location, `Cookie:`) {
			_, _ = s.File.Seek(offsetURLEntry+int64(len(urlSig)), io.SeekStart)
			continue
		}
		locAtParts := strings.SplitN(location, `@`, 2)
		if len(locAtParts) < 2 {
			_, _ = s.File.Seek(offsetURLEntry+int64(len(urlSig)), io.SeekStart)
			continue
		}
		entry.Domain = strings.SplitN(locAtParts[1], `/`, 2)[0]
		directoryIndex, err := bytesx.ReadBytesN(s.File, "directory index", offsetURLEntry, 56, 1)
		if err != nil {
			return nil, err
		}
		entry.DirectoryIndex = directoryIndex
		isCookieEntry := string(entry.DirectoryIndex) == string([]byte{0xFE})
		if !isCookieEntry {
			_, _ = s.File.Seek(offsetURLEntry+int64(len(urlSig)), io.SeekStart)
			continue
		}
		formatVersion, err := bytesx.ReadBytesN(s.File, "entry format version", offsetURLEntry, 58, 1) // 0x00 ⇒ IE5_URL_FILEMAP_ENTRY, 0x10 ⇒ IE6_URL_FILEMAP_ENTRY
		if err != nil {
			return nil, err
		}
		entry.FormatVersion = formatVersion
		offsetURLRecordFileName, err := bytesx.ReadOffSetInt64LE(s.File, "url record filename offset", offsetURLEntry, 60)
		fileName, err := bytesx.ReadString(s.File, "file name", offsetURLEntry, offsetURLRecordFileName)
		if err != nil {
			return nil, err
		}
		entry.FileName = filepath.Join(filepath.Dir(s.FileNameStr), fileName)
		// https://github.com/libyal/libmsiecf/blob/main/documentation/MSIE%20Cache%20File%20(index.dat)%20format.asciidoc#43-cache-entry-flags
		flags, err := bytesx.ReadBytesN(s.File, "flags", offsetURLEntry, 64, 4)
		if err != nil {
			return nil, err
		}
		entry.Flags = binary.LittleEndian.Uint32(flags)
		// probably no Data in Cookie Entries
		offsetURLRecordData, err := bytesx.ReadOffSetInt64LE(s.File, "url record data offset", offsetURLEntry, 68)
		if err != nil {
			return nil, err
		}
		if offsetURLRecordData != 0 {
			urlRecordDataSize, err := bytesx.ReadBytesN(s.File, "url record data size", offsetURLEntry, 72, 4)
			if err != nil {
				return nil, err
			}
			data, err := bytesx.ReadBytesN(s.File, "url record data", offsetURLEntry, offsetURLRecordData, binary.LittleEndian.Uint32(urlRecordDataSize))
			if err != nil {
				return nil, err
			}
			entry.Data = data
		}
		hitsCount, err := bytesx.ReadBytesN(s.File, "hits count", offsetURLEntry, 84, 4)
		if err != nil {
			return nil, err
		}
		entry.HitsCount = binary.LittleEndian.Uint32(hitsCount)

		entries = append(entries, entry)
		textCookieStores = append(
			textCookieStores,
			&TextCookieStore{
				DefaultCookieStore: cookies.DefaultCookieStore{
					BrowserStr:           s.BrowserStr,
					IsDefaultProfileBool: false,
					FileNameStr:          entry.FileName,
				},
			},
		)
		// TODO: the file name of these nested text cookie stores are not visible to the caller, the index.dat appears as the file source

		_, _ = s.File.Seek(offsetURLEntry+int64(len(urlSig)), io.SeekStart)
	}

	var ret []*kooky.Cookie
	for _, textCookieStore := range textCookieStores {
		// TODO: parallelize (internalize kooky/find.go?)
		cs, err := textCookieStore.ReadCookies(filters...)
		if err == nil {
			ret = append(
				ret,
				cs...,
			)
		}
	}

	return ret, nil
}

type CacheCookieEntry struct {
	Domain           string
	FileName         string
	Data             []byte
	DateModification time.Time
	DateLastAccessed time.Time
	DateLastChecked  time.Time
	DateExpiry       time.Time
	FormatVersion    []byte
	BlockCount       int64
	Offset           int64
	DirectoryIndex   []byte
	Flags            uint32
	HitsCount        uint32
}

func scanRest(f io.ReadSeeker, str []byte) (int64, error) {
	index := 0
	r := bufio.NewReader(f)
	offset, _ := f.Seek(0, io.SeekCurrent)
	for index < len(str) {
		b, err := r.ReadByte()
		if err != nil {
			return -1, err
		}
		if str[index] == b {
			index++
		} else {
			index = 0
		}
		offset++
	}
	return offset - int64(len(str)), nil
}

func getFILETIME(r io.ReadSeeker, field string, start, offset int64) (time.Time, error) {
	val, err := bytesx.ReadBytesN(r, field, start, offset, 8)
	if err != nil {
		return time.Time{}, err
	}
	return timex.FromFILETIMEBytes(val)
}

func getFATTIME(r io.ReadSeeker, field string, start, offset int64) (time.Time, error) {
	val, err := bytesx.ReadBytesN(r, field, start, offset, 4)
	if err != nil {
		return time.Time{}, err
	}
	return timex.FromFATTIMEBytes(val)
}

// TODO: filtering with kooky.Domain should search domain by hash or by entry location
// TODO cache v4.7 (IE4), wine index.dat
