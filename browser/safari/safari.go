package safari

// Read safari kooky.Cookie.binarycookies files.
// Thanks to https://github.com/as0ler/BinaryCookieReader

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/cookies"
	"github.com/browserutils/kooky/internal/iterx"
	"github.com/browserutils/kooky/internal/timex"
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

func ReadCookies(ctx context.Context, filename string, filters ...kooky.Filter) ([]*kooky.Cookie, error) {
	return cookies.SingleRead(cookieStore, filename, filters...).ReadAllCookies(ctx)
}

func TraverseCookies(filename string, filters ...kooky.Filter) kooky.CookieSeq {
	return cookies.SingleRead(cookieStore, filename, filters...)
}

func (s *safariCookieStore) TraverseCookies(filters ...kooky.Filter) kooky.CookieSeq {
	return func(yield func(*kooky.Cookie, error) bool) {
		if s == nil {
			yield(nil, errors.New(`cookie store is nil`))
			return
		}
		if err := s.Open(); err != nil {
			yield(nil, err)
			return
		}
		if s.File == nil {
			yield(nil, errors.New(`file is nil`))
			return
		}

		var header fileHeader
		err := binary.Read(s.File, binary.BigEndian, &header)
		if err != nil {
			yield(nil, fmt.Errorf("error reading header: %v", err))
			return
		}
		if string(header.Magic[:]) != "cook" {
			yield(nil, fmt.Errorf("expected first 4 bytes to be %q; got %q", "cook", string(header.Magic[:])))
			return
		}

		pageSizes := make([]int32, header.NumPages)
		if err = binary.Read(s.File, binary.BigEndian, &pageSizes); err != nil {
			yield(nil, fmt.Errorf("error reading page sizes: %w", err))
			return
		}

		// read cookies
		for i, pageSize := range pageSizes {
			if !s.readPage(s.File, i, pageSize, yield, filters...) {
				return
			}
		}

		// TODO(zellyn): figure out how the checksum works.
		var checksum [8]byte
		err = binary.Read(s.File, binary.BigEndian, &checksum)
		if err != nil {
			yield(nil, fmt.Errorf("error reading checksum: %w", err))
			return
		}
	}
}

func (s *safariCookieStore) readPage(f io.Reader, page int, pageSize int32, yield func(*kooky.Cookie, error) bool, filters ...kooky.Filter) bool {
	yld := func(c *kooky.Cookie, e error) bool {
		if e != nil {
			e = fmt.Errorf("error reading page %d: %w", page, e)
		}
		return iterx.CookieFilterYield(context.Background(), c, e, yield, filters...)
	}

	bb := make([]byte, pageSize)
	if _, err := io.ReadFull(f, bb); err != nil {
		return yld(nil, err)
	}
	r := bytes.NewReader(bb)

	var header pageHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return yld(nil, fmt.Errorf("error reading header: %w", err))
	}
	want := [4]byte{0x00, 0x00, 0x01, 0x00}
	if header.Header != want {
		return yld(nil, fmt.Errorf("expected first 4 bytes of page to be %v; got %v", want, header.Header))
	}

	cookieOffsets := make([]int32, header.NumCookies)
	if err := binary.Read(r, binary.LittleEndian, &cookieOffsets); err != nil {
		return yld(nil, fmt.Errorf("error reading cookie offsets: %w", err))
	}

	for i, cookieOffset := range cookieOffsets {
		r.Seek(int64(cookieOffset), io.SeekStart)
		cookie, err := s.readCookie(r)
		if err != nil {
			return yld(nil, fmt.Errorf("cookie %d: %w", i, err))
		}
		if !yld(cookie, nil) {
			return false
		}
	}

	return true
}

func (s *safariCookieStore) readCookie(r io.ReadSeeker) (*kooky.Cookie, error) {
	start, _ := r.Seek(0, io.SeekCurrent)
	var ch cookieHeader
	if err := binary.Read(r, binary.LittleEndian, &ch); err != nil {
		return nil, err
	}

	expiry := timex.FromSafariTime(ch.ExpirationDate)
	creation := timex.FromSafariTime(ch.CreationDate)

	url, err := s.readString(r, "url", start, ch.UrlOffset)
	if err != nil {
		return nil, err
	}
	name, err := s.readString(r, "name", start, ch.NameOffset)
	if err != nil {
		return nil, err
	}
	path, err := s.readString(r, "path", start, ch.PathOffset)
	if err != nil {
		return nil, err
	}
	value, err := s.readString(r, "value", start, ch.ValueOffset)
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
	cookie.Browser = s

	return cookie, nil
}

func (s *safariCookieStore) readString(r io.ReadSeeker, field string, start int64, offset int32) (string, error) {
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

// CookieStore has to be closed with CookieStore.Close() after use.
func CookieStore(filename string, filters ...kooky.Filter) (kooky.CookieStore, error) {
	return cookieStore(filename, filters...)
}

func cookieStore(filename string, filters ...kooky.Filter) (*cookies.CookieJar, error) {
	s := &safariCookieStore{}
	s.FileNameStr = filename
	s.BrowserStr = `safari`

	return cookies.NewCookieJar(s, filters...), nil
}
