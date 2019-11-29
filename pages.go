package parser

import (
	"fmt"
	"io"
)

const (
	TAG_COMMON = 4
)

type Value struct {
	Tag    *Tag
	Buffer []byte
}

func (self *Value) Reader() io.ReaderAt {
	return &BufferReaderAt{self.Buffer}
}

func GetPageValues(ctx *ESEContext, header *PageHeader) []*Value {
	result := []*Value{}

	// Tags are written from the end of the page
	offset := ctx.PageSize + header.Offset - 4

	for tag_count := header.AvailablePageTag(); tag_count > 0; tag_count-- {
		tag := ctx.Profile.Tag(ctx.Reader, offset)
		value_offset := header.Offset + 40 + int64(tag.ValueOffset())

		buffer := make([]byte, int(tag.ValueSize()))
		ctx.Reader.ReadAt(buffer, value_offset)

		result = append(result, &Value{Tag: tag, Buffer: buffer})
		offset -= 4
	}

	return result
}

func GetRoot(ctx *ESEContext, value *Value) *ESENT_ROOT_HEADER {
	return ctx.Profile.ESENT_ROOT_HEADER(value.Reader(), 0)
}

func GetBranch(ctx *ESEContext, value *Value) *ESENT_BRANCH_HEADER {
	return ctx.Profile.ESENT_BRANCH_HEADER(value.Reader(), 0)
}

func (self *PageHeader) IsBranch() bool {
	return !self.Flags().IsSet("Leaf")
}

func (self *PageHeader) IsLeaf() bool {
	return self.Flags().IsSet("Leaf")
}

func DumpPage(ctx *ESEContext, id int64) {
	header := ctx.GetPage(id)
	fmt.Printf("Page %v: %v\n", id, header.DebugString())

	// Show the tags
	values := GetPageValues(ctx, header)
	if len(values) == 0 {
		return
	}

	for _, value := range values {
		fmt.Println(value.Tag.DebugString())
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
			&BufferReaderAt{values[0].Buffer}, 0).Dump()

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
	if value.Tag.Flags()&TAG_COMMON > 0 {
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
	if value.Tag.Flags()&TAG_COMMON > 0 {
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
	if id <= 0 {
		return nil
	}

	header := ctx.GetPage(id)
	values := GetPageValues(ctx, header)

	// No more records.
	if len(values) == 0 {
		return nil
	}

	for _, value := range values[1:] {
		if header.IsLeaf() {
			// Allow the callback to return early (e.g. in case of cancellation)
			err := cb(header, id, value)
			if err != nil {
				return err
			}
		} else if header.IsBranch() {
			// Walk the branch
			branch := NewESENT_BRANCH_ENTRY(ctx, value)
			err := WalkPages(ctx, branch.ChildPageNumber(), cb)
			if err != nil {
				return err
			}
		}
	}

	if header.NextPageNumber() > 0 {
		err := WalkPages(ctx, int64(header.NextPageNumber()), cb)
		if err != nil {
			return err
		}
	}

	return nil
}
