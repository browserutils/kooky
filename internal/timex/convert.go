package timex

import (
	"encoding/binary"
	"errors"
	"math"
	"time"
)

func FromFILETIME(timestamp_utc int64) time.Time {
	// https://github.com/golang/go/blob/5d1a95175e693f5be0bc31ae9e6a7873318925eb/src/syscall/types_windows.go#L352
	timestamp_utc -= 116444736e9
	return time.Unix(0, timestamp_utc*100)
}

func FromFILETIMESplit[T ~[4]byte | ~uint32 | ~int32 | ~uint64 | ~int64](low, high T) time.Time {
	var lowInt64, highInt64 int64
	switch lowTyp := any(low).(type) {
	case [4]byte:
		highTyp := [4]byte(any(high).([4]byte))
		lowInt64 = int64(binary.LittleEndian.Uint32(lowTyp[:]))
		highInt64 = int64(binary.LittleEndian.Uint32(highTyp[:]))
	case uint32:
		lowInt64 = int64(lowTyp)
		highInt64 = int64(any(high).(uint32))
	case int32:
		lowInt64 = int64(lowTyp)
		highInt64 = int64(any(high).(int32))
	case uint64:
		lowInt64 = int64(lowTyp)
		highInt64 = int64(any(high).(uint64))
	case int64:
		lowInt64 = lowTyp
		highInt64 = any(high).(int64)
	}
	// https://github.com/golang/go/blob/5d1a95175e693f5be0bc31ae9e6a7873318925eb/src/syscall/types_windows.go#L352
	timestamp_utc := int64(highInt64)<<32 + int64(lowInt64)
	return FromFILETIME(timestamp_utc)
}

func FromFILETIMEBytes(val []byte) (time.Time, error) {
	if len(val) != 8 {
		return time.Time{}, errors.New(`length of byte slice is not 8`)
	}
	return FromFILETIMESplit(*(*[4]byte)(val[:4]), *(*[4]byte)(val[4:8])), nil
}

func FromFATTIMEBytes(val []byte) (time.Time, error) {
	if len(val) != 4 {
		return time.Time{}, errors.New(`length of byte slice is not 4`)
	}
	// https://github.com/libyal/libmsiecf/blob/main/documentation/MSIE%20Cache%20File%20(index.dat)%20format.asciidoc#fat_date_time
	date := binary.LittleEndian.Uint16(val[:2])
	tm := binary.LittleEndian.Uint16(val[2:])
	return time.Date(1980+int(date>>9), time.Month((date>>5)&0b00001111), int(date&0b00011111), int(tm>>11), int((tm>>5)&0b00111111), 2*int(tm&0b00011111), 0, time.Now().Location()), nil
}

// FromSafariTime converts double seconds to a time.Time object,
// accounting for the switch to Mac epoch (Jan 1 2001).
func FromSafariTime(floatSecs float64) time.Time {
	seconds, frac := math.Modf(floatSecs)
	return time.Unix(int64(seconds)+978307200, int64(frac*1000000000))
}
