package utils

import "time"

func FromFILETIME(timestamp_utc int64) time.Time {
	// https://github.com/golang/go/blob/5d1a95175e693f5be0bc31ae9e6a7873318925eb/src/syscall/types_windows.go#L352
	timestamp_utc -= 116444736e9
	return time.Unix(0, timestamp_utc*100)
}
