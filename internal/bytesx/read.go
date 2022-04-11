package bytesx

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

type intNumber interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func ReadString[S, T intNumber](r io.ReadSeeker, field string, start S, offset T) (string, error) {
	if _, err := r.Seek(int64(start)+int64(offset), io.SeekStart); err != nil {
		return "", fmt.Errorf("seeking for %q at offset %d", field, offset)
	}
	b := bufio.NewReader(r)
	value, err := b.ReadString(0)
	if err != nil {
		return "", fmt.Errorf("reading for %q at offset %d", field, offset)
	}

	return value[:len(value)-1], nil
}

func ReadBytesN[R, S, T intNumber](r io.ReadSeeker, field string, start S, offset T, length R) ([]byte, error) {
	if _, err := r.Seek(int64(start)+int64(offset), io.SeekStart); err != nil {
		return nil, fmt.Errorf("seeking for %q at offset %d", field, offset)
	}
	val := make([]byte, length)
	if _, err := r.Read(val); err != nil {
		return nil, fmt.Errorf("reading for %q at offset %d", field, offset)
	}
	return val, nil
}

func ReadOffSetInt64LE[S, T intNumber](r io.ReadSeeker, field string, start S, offset T) (int64, error) {
	if _, err := r.Seek(int64(start)+int64(offset), io.SeekStart); err != nil {
		return 0, fmt.Errorf("seeking for %q at offset %d", field, offset)
	}
	val := make([]byte, 4)
	if _, err := r.Read(val); err != nil {
		return 0, fmt.Errorf("reading for %q at offset %d", field, offset)
	}
	return int64(binary.LittleEndian.Uint32(val)), nil
}
