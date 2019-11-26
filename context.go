package parser

import (
	"io"
)

type ESEContext struct {
	Reader   io.ReaderAt
	FileSize int64
	Profile  *ESEProfile
	PageSize int64
	Header   *FileHeader
}

func NewESEContext(reader io.ReaderAt, filesize int64) (*ESEContext, error) {
	result := &ESEContext{
		Profile: NewESEProfile(),
		Reader:  reader,
	}

	// TODO error check.
	result.Header = result.Profile.FileHeader(reader, 0)
	result.PageSize = int64(result.Header.PageSize())

	return result, nil
}

func (self *ESEContext) GetPage(id int64) *PageHeader {
	// First file page is file header, second page is backup of file header.
	result := self.Profile.PageHeader(self.Reader, (id+1)*self.PageSize)
	return result
}
