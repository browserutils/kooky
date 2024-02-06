package parser

import "io"

type BufferReaderAt struct {
	buffer []byte
}

func (self *BufferReaderAt) ReadAt(buf []byte, offset int64) (int, error) {
	to_read := int64(len(buf))
	if offset < 0 {
		to_read += offset
		offset = 0
	}

	if offset+to_read > int64(len(self.buffer)) {
		to_read = int64(len(self.buffer)) - offset
	}

	if to_read < 0 {
		return 0, nil
	}

	n := copy(buf, self.buffer[offset:offset+to_read])

	return n, nil
}

type OffsetReader struct {
	reader io.ReaderAt
	offset int64
	length int64
}

func (self OffsetReader) ReadAt(buff []byte, off int64) (int, error) {
	to_read := int64(len(buff))
	if off+to_read > self.length {
		to_read = self.length - off
	}

	if to_read < 0 {
		return 0, nil
	}
	return self.reader.ReadAt(buff, off+self.offset)
}

func NewOffsetReader(reader io.ReaderAt, offset, size int64) io.ReaderAt {
	return &OffsetReader{
		reader: reader,
		offset: offset,
		length: offset + size,
	}
}
