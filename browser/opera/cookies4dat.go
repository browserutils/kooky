package opera

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"math/bits"
	"strconv"
	"time"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/internal/iterx"
)

type fileHeader struct {
	FileVersionNumber uint32
	AppVersionNumber  uint32
	IDTagLength       uint16
	LengthLength      uint16
}

type record struct {
	TagIDType         any
	PayloadLengthType any
	Payload           []byte
}

// "cookies4.dat" format
func (s *operaPrestoCookieStore) TraverseCookies(filters ...kooky.Filter) kooky.CookieSeq {
	if s == nil {
		return iterx.ErrCookieSeq(errors.New(`cookie store is nil`))
	}
	return func(yield func(*kooky.Cookie, error) bool) {
		if err := s.Open(); err != nil {
			yield(nil, err)
			return
		}
		if _, err := s.File.Seek(0, io.SeekStart); err != nil {
			yield(nil, err)
			return
		}

		var hdr fileHeader
		if err := binary.Read(s.File, binary.BigEndian, &hdr); err != nil {
			yield(nil, err)
			return
		}
		fileFormatVersionMajor := hdr.FileVersionNumber >> 12
		fileFormatVersionMinor := hdr.FileVersionNumber & 0xfff
		if fileFormatVersionMajor != 1 || fileFormatVersionMinor != 0 {
			yield(nil, errors.New(`unsupported file format version `+
				strconv.Itoa(int(fileFormatVersionMajor))+`.`+strconv.Itoa(int(fileFormatVersionMinor))))
			return
		}
		// appVersionMajor := hdr.AppVersionNumber >> 12
		// appVersionMinor := hdr.AppVersionNumber & 0xfff

		p := &processor{
			browser:      s,
			reader:       s.File,
			idTagLength:  hdr.IDTagLength,
			lengthLength: hdr.LengthLength,
			filters:      filters,
		}
		yld := func(cookie *kooky.Cookie, err error) bool {
			return iterx.CookieFilterYield(context.Background(), cookie, err, yield, filters...)
		}
		_, err := p.process(yld)
		if err != nil {
			if !p.end && !errors.Is(err, iterx.ErrYieldEnd) && !errors.Is(err, io.EOF) {
				_ = yld(nil, err)
			}
			return
		}
		if p.end || !yld(p.cookie, nil) {
			return
		}
	}
}

type processor struct {
	reader        io.Reader
	idTagLength   uint16
	lengthLength  uint16
	tagID         uint32
	payloadLength uint32
	domainParts   []string
	path          string
	cookie        *kooky.Cookie
	filters       []kooky.Filter
	browser       kooky.BrowserInfo
	end           bool
}

func (p *processor) process(yield func(*kooky.Cookie, error) bool) (int, error) {
	if p.idTagLength < 1 || p.idTagLength > 4 || p.lengthLength < 1 || p.lengthLength > 4 {
		err := errors.New(`unexpected byte length values`)
		if !yield(nil, err) {
			p.end = true
		}
		return 0, err
	}

	n, tagID, payloadLength, err := getRecord(p.reader, p.idTagLength, p.lengthLength)
	isEOF := errors.Is(err, io.EOF)
	if isEOF {
		p.tagID = tagID
		p.payloadLength = payloadLength
	}
	if err != nil {
		return n, err
	}
	p.tagID = tagID
	p.payloadLength = payloadLength

	var payload []byte
	if payloadLength > 0 {
		switch p.tagID {
		case tagIDDomainStart, tagIDPathStart, tagIDCookie:
		default:
			payload = make([]byte, payloadLength)
			n2, err := p.reader.Read(payload)
			n += n2
			if err != nil {
				return n, err
			}
		}
	}
	switch tagID {
	case tagIDCookie:
		c := &kooky.Cookie{}
		var domain string
		for i := 0; i < len(p.domainParts); i++ {
			if i == 0 {
				domain = p.domainParts[len(p.domainParts)-i-1]
			} else {
				domain += `.` + p.domainParts[len(p.domainParts)-i-1]
			}
		}
		c.Domain = domain
		c.Path = p.path
		c.Browser = p.browser
		if !yield(p.cookie, nil) {
			p.end = true
			return n, err
		}
		p.cookie = c
	case tagIDDomainName:
		p.domainParts = append(p.domainParts, string(payload))
	case tagIDDomainEnd:
		if len(p.domainParts) > 0 {
			p.domainParts = p.domainParts[:len(p.domainParts)-1]
		}
	case tagIDPathName:
		p.path = string(payload)
	case tagIDPathStart, tagIDPathEnd:
		p.path = ``
	case tagIDCookieName:
		p.cookie.Name = string(payload)
	case tagIDCookieValue:
		p.cookie.Value = string(payload)
	case tagIDCookieDateExpiry:
		if len(payload) != 8 {
			return n, err
		}
		p.cookie.Expires = time.Unix(int64(binary.BigEndian.Uint64(payload)), 0)
	case tagIDCookieHTTPSOnly:
		p.cookie.Secure = true
	}

	if !isEOF {
		var n3 int
		n3, err = p.process(yield)
		n += n3
	}

	return n, err
}

func getRecord(s io.Reader, idTagLength, lengthLength uint16) (n int, tagID, payloadLength uint32, e error) {
	tagIDBytes := make([]byte, idTagLength)
	m, err := s.Read(tagIDBytes)
	n += m
	if err != nil {
		return n, 0, 0, err
	}

	noLength := bits.LeadingZeros8(tagIDBytes[0]) == 0
	if noLength {
		tagIDBytes[0] &= 0x7f // remove most significant bit
	}
	tagID = toUint32(tagIDBytes)
	if noLength {
		return n, tagID, 0, nil
	}

	payloadLengthBytes := make([]byte, lengthLength)
	m, err = s.Read(payloadLengthBytes)
	n += m
	if err != nil {
		return n, 0, 0, err
	}
	payloadLength = toUint32(payloadLengthBytes)

	return n, tagID, payloadLength, nil
}

func toUint32(b []byte) uint32 {
	if l := len(b); l > 4 || l < 1 {
		panic(`unexpected byte length values`)
	}
	length := len(b)
	switch length {
	case 1:
		b = append([]byte{0x00, 0x00, 0x00}, b...)
	case 2:
		b = append([]byte{0x00, 0x00}, b...)
	case 3:
		b = append([]byte{0x00}, b...)
	}
	return binary.BigEndian.Uint32(b)
}

const (
	tagIDDomainStart                         uint32 = 0x01 // struct
	tagIDDomainEnd                           uint32 = 0x04 // --
	tagIDPathStart                           uint32 = 0x02 // struct
	tagIDPathEnd                             uint32 = 0x05 // --
	tagIDCookie                              uint32 = 0x03 // struct
	tagIDDomainName                          uint32 = 0x1e // string
	tagIDDomainFilter                        uint32 = 0x1f // int8
	tagIDDomainPathFilter                    uint32 = 0x21 // int8
	tagIDDomain3rdPartyFilter                uint32 = 0x25 // int8
	tagIDPathName                            uint32 = 0x1d // string
	tagIDCookieName                          uint32 = 0x10 // string
	tagIDCookieValue                         uint32 = 0x11 // string
	tagIDCookieDateExpiry                    uint32 = 0x12 // time_t
	tagIDCookieDateLastUsed                  uint32 = 0x13 // time_t
	tagIDCookieRFC2965Comment                uint32 = 0x14 // string
	tagIDCookieRFC2965CommentURL             uint32 = 0x15 // string
	tagIDCookieRFC2965CommentVersion1Domain  uint32 = 0x16 // string
	tagIDCookieRFC2965CommentVersion1Path    uint32 = 0x17 // string
	tagIDCookieRFC2965CommentVersion1PortLim uint32 = 0x18 // string
	tagIDCookieHTTPSOnly                     uint32 = 0x19 // flag
	tagIDCookieRFC2965Version                uint32 = 0x1a // int8+
	tagIDCookieOnlyToSource                  uint32 = 0x1b // flag
	tagIDCookieDeleteProtection              uint32 = 0x1c // flag - never implemented by Opera
	tagIDCookiePathPrefixFilter              uint32 = 0x20 // flag
	tagIDCookiePasswordLogin                 uint32 = 0x22 // flag
	tagIDCookieHTTPAuth                      uint32 = 0x23 // flag
	tagIDCookie3rdParty                      uint32 = 0x24 // flag
)
