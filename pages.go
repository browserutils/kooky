package parser

import (
	"fmt"
	"io"
)

const (
	TAG_COMMON = 4
)

// TODO: This is called LINE in the MS code. It represents a single
// node in the page. Depending on the page type it needs to be further
// parsed.
type Value struct {
	Tag    *Tag
	PageID int64
	Flags  uint64

	reader       io.ReaderAt
	BufferOffset int64
	BufferSize   int64
}

func (self *Value) GetBuffer() []byte {
	result := make([]byte, self.BufferSize)
	self.reader.ReadAt(result, self.BufferOffset)
	return result
}

func (self *Value) Reader() io.ReaderAt {
	return NewOffsetReader(self.reader,
		self.BufferOffset, self.BufferSize)
}

func NewReaderValue(ctx *ESEContext, tag *Tag, PageID int64,
	reader io.ReaderAt, start, length int64) *Value {
	result := &Value{Tag: tag, PageID: PageID, reader: reader,
		BufferOffset: start, BufferSize: length}
	if ctx.Version == 0x620 && ctx.Revision >= 17 &&
		ctx.PageSize > 8192 && length > 0 {

		buffer := make([]byte, 4)
		reader.ReadAt(buffer, start)

		result.Flags = uint64(buffer[1] >> 5)
		buffer[1] &= 0x1f
	} else {
		result.Flags = uint64(tag._ValueOffset()) >> 13
	}
	return result
}

func (self *Tag) valueOffset(ctx *ESEContext) uint16 {
	if ctx.Version == 0x620 && ctx.Revision >= 17 && ctx.PageSize > 8192 {
		return self._ValueOffset() & 0x7FFF
	}
	return self._ValueOffset() & 0x1FFF
}

func (self *Tag) ValueOffsetInPage(ctx *ESEContext, page *PageHeader) int64 {
	return int64(self.valueOffset(ctx)) + page.EndOffset(ctx)
}

func (self *Tag) FFlags() uint16 {
	// CPAGE::TAG::FFlags
	// https://github.com/microsoft/Extensible-Storage-Engine/blob/933dc839b5a97b9a5b3e04824bdd456daf75a57d/dev/ese/src/ese/cpage.cxx#1212
	return (self._ValueOffset() & 0x1fff) >> 13
}

func (self *Tag) ValueSize(ctx *ESEContext) uint16 {
	if ctx.Version == 0x620 && ctx.Revision >= 17 &&
		!IsSmallPage(ctx.PageSize) {
		return self._ValueSize() & 0x7FFF
	}
	return self._ValueSize() & 0x1FFF
}

func GetPageValues(ctx *ESEContext, header *PageHeader, id int64) []*Value {
	result := []*Value{}

	// Tags are written from the end of the page. Sizeof(Tag) = 4
	offset := ctx.PageSize + header.Offset - 4

	// Skip the external value tag because it is fetched using a
	// dedicated call to PageHeader.ExternalValue()
	offset -= 4

	for tag_count := header.AvailablePageTag(); tag_count > 0; tag_count-- {
		tag := ctx.Profile.Tag(ctx.Reader, offset)

		result = append(result, NewReaderValue(
			ctx, tag, id, ctx.Reader,
			tag.ValueOffsetInPage(ctx, header),
			int64(tag.ValueSize(ctx))))
		offset -= 4
	}

	if DebugWalk {
		fmt.Printf("Got %v values for page %v\n", len(result), id)
	}
	return result
}

func GetRoot(ctx *ESEContext, value *Value) *ESENT_ROOT_HEADER {
	return ctx.Profile.ESENT_ROOT_HEADER(value.Reader(), 0)
}

func GetBranch(ctx *ESEContext, value *Value) *ESENT_BRANCH_HEADER {
	return ctx.Profile.ESENT_BRANCH_HEADER(value.Reader(), 0)
}

type PageHeader struct {
	*PageHeader_

	// The value pointed to by tag 0
	external_value_bytes []byte
}

func (self *PageHeader) ExternalValueBytes(ctx *ESEContext) []byte {
	if self.external_value_bytes != nil {
		return self.external_value_bytes
	}

	self.external_value_bytes = self.ExternalValue(ctx).GetBuffer()
	return self.external_value_bytes
}

// The External value is the zero'th tag
func (self *PageHeader) ExternalValue(ctx *ESEContext) *Value {
	offset := ctx.PageSize + self.Offset - 4
	tag := self.Profile.Tag(self.Reader, offset)

	return NewReaderValue(
		ctx, tag, 0, ctx.Reader,
		tag.ValueOffsetInPage(ctx, self),
		int64(tag.ValueSize(ctx)))
}

func (self *PageHeader) IsBranch() bool {
	return !self.Flags().IsSet("Leaf")
}

func (self *PageHeader) IsLeaf() bool {
	return self.Flags().IsSet("Leaf")
}

func (self *PageHeader) EndOffset(ctx *ESEContext) int64 {
	size := int64(40)

	// The header is larger when the pagesize is bigger (PGHDR2 vs
	// PGHDR)
	// https://github.com/microsoft/Extensible-Storage-Engine/blob/933dc839b5a97b9a5b3e04824bdd456daf75a57d/dev/ese/src/inc/cpage.hxx#L885
	if !IsSmallPage(ctx.PageSize) {
		size = 80
	}
	return self.Offset + size
}

func DumpPage(ctx *ESEContext, id int64) {
	header := ctx.GetPage(id)
	fmt.Printf("Page %v: %v\n", id, header.DebugString())

	// Show the tags
	values := GetPageValues(ctx, header, id)
	if len(values) == 0 {
		return
	}

	for i, value := range values {
		fmt.Printf("Tag %v @ %#x offset %#x length %#x\n",
			i, value.Tag.Offset,
			value.Tag.ValueOffsetInPage(ctx, header),
			value.Tag.ValueSize(ctx))
	}

	flags := header.Flags()

	if flags.IsSet("Root") {
		GetRoot(ctx, values[0]).Dump()

		// Branch header
	} else if header.IsBranch() {
		GetBranch(ctx, values[0]).Dump()

		// SpaceTree header
	} else if flags.IsSet("SpaceTree") {
		ctx.Profile.ESENT_SPACE_TREE_HEADER(
			ctx.Reader, values[0].BufferOffset).Dump()

		// Leaf header
	} else if header.IsLeaf() {
		NewESENT_LEAF_ENTRY(ctx, values[0]).Dump()
	}

	for _, value := range values[1:] {
		if header.IsBranch() {
			NewESENT_BRANCH_ENTRY(ctx, value).Dump()
		} else if header.IsLeaf() {
			if flags.IsSet("SpaceTree") {
				ctx.Profile.ESENT_SPACE_TREE_ENTRY(value.Reader(), 0).Dump()
			} else if flags.IsSet("Index") {
				ctx.Profile.ESENT_INDEX_ENTRY(value.Reader(), 0).Dump()
			} else if flags.IsSet("Long") {
				// TODO
			} else {
				NewESENT_LEAF_ENTRY(ctx, value).Dump()
			}
		}
	}
}

func (self *ESENT_ROOT_HEADER) Dump() {
	fmt.Println(self.DebugString())
}

func (self *ESENT_SPACE_TREE_ENTRY) Dump() {
	fmt.Println(self.DebugString())
}
func (self *ESENT_INDEX_ENTRY) Dump() {
	fmt.Println(self.DebugString())
}

// NewESENT_LEAF_ENTRY creates a new ESENT_LEAF_ENTRY
// object. Depending on the Tag flags, there may be present a
// CommonPageKeySize field before the struct. This constructor then
// positions the struct appropriately.
func NewESENT_LEAF_ENTRY(ctx *ESEContext, value *Value) *ESENT_LEAF_ENTRY {
	if value.Flags&TAG_COMMON > 0 {
		// Skip the common header
		return ctx.Profile.ESENT_LEAF_ENTRY(value.Reader(), 2)
	}
	return ctx.Profile.ESENT_LEAF_ENTRY(value.Reader(), 0)
}

func (self *ESENT_LEAF_ENTRY) Dump() {
	fmt.Println(self.DebugString())
}

func (self *ESENT_LEAF_ENTRY) EntryData() int64 {
	// Tag includes Local Page Key - skip it and the common page key
	return self.Offset + 2 + int64(self.LocalPageKeySize())
}

func (self *ESENT_BRANCH_HEADER) Dump() {
	fmt.Println(self.DebugString())
}

// NewESENT_BRANCH_ENTRY creates a new ESENT_BRANCH_ENTRY
// object. Depending on the Tag flags, there may be present a
// CommonPageKeySize field before the struct. This construstor then
// positions the struct appropriately.
func NewESENT_BRANCH_ENTRY(ctx *ESEContext, value *Value) *ESENT_BRANCH_ENTRY {
	if value.Flags&TAG_COMMON > 0 {
		// Skip the common header
		return ctx.Profile.ESENT_BRANCH_ENTRY(value.Reader(), 2)
	}
	return ctx.Profile.ESENT_BRANCH_ENTRY(value.Reader(), 0)
}

func (self *ESENT_BRANCH_ENTRY) Dump() {
	fmt.Printf("%s", self.DebugString())
	fmt.Printf("  ChildPageNumber: %#x\n", self.ChildPageNumber())
}

func (self *ESENT_BRANCH_ENTRY) ChildPageNumber() int64 {
	return int64(ParseUint32(self.Reader, self.Offset+2+
		int64(self.LocalPageKeySize())))
}

func (self *ESENT_SPACE_TREE_HEADER) Dump() {
	fmt.Println(self.DebugString())
}

func (self *ESENT_LEAF_HEADER) Dump() {
	fmt.Println(self.DebugString())
}

// WalkPages walks the b tree starting with the page id specified and
// extracts all tagged values into the callback. The callback may
// return an error which will cause WalkPages to stop and relay that
// error to our caller.
func WalkPages(ctx *ESEContext,
	id int64,
	cb func(header *PageHeader, page_id int64, value *Value) error) error {
	seen := make(map[int64]bool)

	return _walkPages(ctx, id, seen, cb)
}

func _walkPages(ctx *ESEContext,
	id int64, seen map[int64]bool,
	cb func(header *PageHeader, page_id int64, value *Value) error) error {

	_, pres := seen[id]
	if id <= 0 || pres {
		return nil
	}
	seen[id] = true

	header := ctx.GetPage(id)
	values := GetPageValues(ctx, header, id)
	if DebugWalk {
		fmt.Printf("Walking page %v %v\n", id, header.DebugString())
	}

	// No more records.
	if len(values) == 0 {
		return nil
	}

	for _, value := range values {
		if header.IsLeaf() {
			// Allow the callback to return early (e.g. in case of
			// cancellation)
			err := cb(header, id, value)
			if err != nil {
				return err
			}
		} else if header.IsBranch() {
			// Walk the branch
			branch := NewESENT_BRANCH_ENTRY(ctx, value)
			err := _walkPages(ctx, branch.ChildPageNumber(), seen, cb)
			if err != nil {
				return err
			}
		}
	}

	if header.NextPageNumber() > 0 {
		err := _walkPages(ctx, int64(header.NextPageNumber()), seen, cb)
		if err != nil {
			return err
		}
	}

	return nil
}
