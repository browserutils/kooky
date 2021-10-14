package parser

import (
	"io"
	"time"
)

func WinFileTime64(reader io.ReaderAt, offset int64) time.Time {
	value := ParseInt64(reader, offset)
	return time.Unix((value/10000000)-11644473600, 0).UTC()
}
