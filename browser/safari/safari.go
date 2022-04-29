package safari

// Read safari kooky.Cookie.binarycookies files.
// Thanks to https://github.com/as0ler/BinaryCookieReader

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal/cookies"
	"github.com/zellyn/kooky/internal/timex"
)

type fileHeader struct {
	Magic    [4]byte
	NumPages int32
}

type pageHeader struct {
	Header     [4]byte
	NumCookies int32
}

type cookieHeader struct {
	Size           int32
	Unknown1       int32
	Flags          int32
	Unknown2       int32
	UrlOffset      int32
	NameOffset     int32
	PathOffset     int32
	ValueOffset    int32
	End            [8]byte
	ExpirationDate float64
	CreationDate   float64
}

type safariCookieStore struct {
	cookies.DefaultCookieStore
}

var _ cookies.CookieStore = (*safariCookieStore)(nil)

func ReadCookies(filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	s, err := cookieStore(filename, filters...)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	return s.ReadCookies(filters...)
}

func (s *safariCookieStore) ReadCookies(filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	if s == nil {
		return nil, errors.New(`cookie store is nil`)
	}
	if err := s.Open(); err != nil {
		return nil, err
	} else if s.File == nil {
		return nil, errors.New(`file is nil`)
	}

	var allCookies []*kooky.Cookie

	var header fileHeader
	err := binary.Read(s.File, binary.BigEndian, &header)
	if err != nil {
		return nil, fmt.Errorf("error reading header: %v", err)
	}
	if string(header.Magic[:]) != "cook" {
		return nil, fmt.Errorf("expected first 4 bytes to be %q; got %q", "cook", string(header.Magic[:]))
	}

	pageSizes := make([]int32, header.NumPages)
	if err = binary.Read(s.File, binary.BigEndian, &pageSizes); err != nil {
		return nil, fmt.Errorf("error reading page sizes: %v", err)
	}

	for i, pageSize := range pageSizes {
		if allCookies, err = readPage(s.File, pageSize, allCookies); err != nil {
			return nil, fmt.Errorf("error reading page %d: %v", i, err)
		}
	}

	// TODO(zellyn): figure out how the checksum works.
	var checksum [8]byte
	err = binary.Read(s.File, binary.BigEndian, &checksum)
	if err != nil {
		return nil, fmt.Errorf("error reading checksum: %v", err)
	}

	// Filter cookies by specified filters.
	cookies := kooky.FilterCookies(allCookies, filters...)

	return cookies, nil
}

func readPage(f io.Reader, pageSize int32, cookies []*kooky.Cookie) ([]*kooky.Cookie, error) {
	bb := make([]byte, pageSize)
	if _, err := io.ReadFull(f, bb); err != nil {
		return nil, err
	}
	r := bytes.NewReader(bb)

	var header pageHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("error reading header: %v", err)
	}
	want := [4]byte{0x00, 0x00, 0x01, 0x00}
	if header.Header != want {
		return nil, fmt.Errorf("expected first 4 bytes of page to be %v; got %v", want, header.Header)
	}

	cookieOffsets := make([]int32, header.NumCookies)
	if err := binary.Read(r, binary.LittleEndian, &cookieOffsets); err != nil {
		return nil, fmt.Errorf("error reading cookie offsets: %v", err)
	}

	for i, cookieOffset := range cookieOffsets {
		r.Seek(int64(cookieOffset), io.SeekStart)
		cookie, err := readCookie(r)
		if err != nil {
			return nil, fmt.Errorf("cookie %d: %v", i, err)
		}
		cookies = append(cookies, cookie)
	}

	return cookies, nil
}

func readCookie(r io.ReadSeeker) (*kooky.Cookie, error) {
	start, _ := r.Seek(0, io.SeekCurrent)
	var ch cookieHeader
	if err := binary.Read(r, binary.LittleEndian, &ch); err != nil {
		return nil, err
	}

	expiry := timex.FromSafariTime(ch.ExpirationDate)
	creation := timex.FromSafariTime(ch.CreationDate)

	url, err := readString(r, "url", start, ch.UrlOffset)
	if err != nil {
		return nil, err
	}
	name, err := readString(r, "name", start, ch.NameOffset)
	if err != nil {
		return nil, err
	}
	path, err := readString(r, "path", start, ch.PathOffset)
	if err != nil {
		return nil, err
	}
	value, err := readString(r, "value", start, ch.ValueOffset)
	if err != nil {
		return nil, err
	}

	cookie := &kooky.Cookie{}

	cookie.Expires = expiry
	cookie.Creation = creation
	cookie.Name = name
	cookie.Value = value
	cookie.Domain = url
	cookie.Path = path
	cookie.Secure = (ch.Flags & 1) > 0
	cookie.HttpOnly = (ch.Flags & 4) > 0
	return cookie, nil
}

func readString(r io.ReadSeeker, field string, start int64, offset int32) (string, error) {
	if _, err := r.Seek(start+int64(offset), io.SeekStart); err != nil {
		return "", fmt.Errorf("seeking for %q at offset %d", field, offset)
	}
	b := bufio.NewReader(r)
	value, err := b.ReadString(0)
	if err != nil {
		return "", fmt.Errorf("reading for %q at offset %d", field, offset)
	}

	return value[:len(value)-1], nil
}

// CookieJar returns an initiated http.CookieJar based on the cookies stored by
// the Safari browser. Set cookies are memory stored and do not modify any
// browser files.
//
func CookieJar(filename string, filters ...kooky.Filter) (http.CookieJar, error) {
	j, err := cookieStore(filename, filters...)
	if err != nil {
		return nil, err
	}
	defer j.Close()
	if err := j.InitJar(); err != nil {
		return nil, err
	}
	return j, nil
}

// CookieStore has to be closed with CookieStore.Close() after use.
//
func CookieStore(filename string, filters ...kooky.Filter) (kooky.CookieStore, error) {
	return cookieStore(filename, filters...)
}

func cookieStore(filename string, filters ...kooky.Filter) (*cookies.CookieJar, error) {
	s := &safariCookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `safari`

	return &cookies.CookieJar{CookieStore: s}, nil
}
