package parser

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

const (
	// Flags on the record header
	COMPRESSION      = 0x02
	INLINE_STRING    = 0x01
	INLINE_STRING_2  = 0x08
	LZMA_COMPRESSION = 0x18 // Not supported
)

func Decompress7BitCompression(buf []byte) string {
	result := make([]byte, 0, (len(buf)+5)*8/7)

	value_16bit := uint16(0)
	bit_index := 0

	for i := 1; i < len(buf); i++ {
		slice := buf[i : i+1]
		if i+1 < len(buf) {
			slice = append(slice, buf[i+1])
		}

		for len(slice) < 2 {
			slice = append(slice, 0)
		}

		value_16bit |= binary.LittleEndian.Uint16(slice) << bit_index
		result = append(result, byte(value_16bit&0x7f))

		value_16bit >>= 7
		bit_index++

		if bit_index == 7 {
			result = append(result, byte(value_16bit&0x7f))
			value_16bit >>= 7
			bit_index = 0
		}
	}

	return strings.Split(string(result), "\x00")[0]
}

func ParseLongText(buf []byte, flag uint32) string {
	if len(buf) < 2 {
		return ""
	}

	// fmt.Printf("Column Flags %v\n", flag)
	leading_byte := buf[0]
	if leading_byte != 0 && leading_byte != 1 && leading_byte != 8 &&
		leading_byte != 3 && leading_byte != 0x18 {
		return strings.Split(
			UTF16BytesToUTF8(buf, binary.LittleEndian), "\x00")[0]

	}
	// fmt.Printf("Inline Flags %v\n", flag)

	// Lzxpress compression - not supported right now.
	if leading_byte == 0x18 {
		fmt.Printf("LZXPRESS compression not supported currently\n")
		return strings.Split(string(buf), "\x00")[0]
	}

	// The following is either 7 bit compressed or utf16 encoded. Its
	// hard to figure out which it is though because there is no
	// consistency in the flags. We do our best to guess!!
	var result string
	if len(buf) >= 3 && buf[2] == 0 {
		// Probably UTF16 encoded
		result = UTF16BytesToUTF8(buf[1:], binary.LittleEndian)

	} else {
		// Probably 7bit compressed
		result = Decompress7BitCompression(buf[1:])
	}

	//fmt.Printf("returned %v\n", result)
	return strings.Split(result, "\x00")[0]
}

func ParseText(reader io.ReaderAt, offset int64, len int64, flags uint32) string {
	if len < 0 {
		return ""

	}
	if len > 1024*10 {
		len = 1024 * 10
	}

	data := make([]byte, len)
	n, err := reader.ReadAt(data, offset)
	if err != nil {
		return ""
	}
	data = data[:n]

	var str string
	if flags == 1 {
		str = string(data[:n])
	} else {
		str = UTF16BytesToUTF8(data, binary.LittleEndian)
	}
	return strings.Split(str, "\x00")[0]
}
