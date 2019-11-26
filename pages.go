package parser

import (
	"errors"
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

func RootNodeVisitor(ctx *ESEContext, id int64) error {
	fmt.Printf("RootNodeVisitor: %v\n", id)

	root_page := ctx.GetPage(id)
	if !root_page.Flags().IsSet("Root") {
		return errors.New(fmt.Sprintf("ID %v: Not root node", id))
	}

	values := GetPageValues(ctx, root_page)
	if len(values) == 0 {
		return nil
	}

	reader := &BufferReaderAt{values[0].Buffer}
	root_header := ctx.Profile.ESENT_ROOT_HEADER(reader, 0)

	if root_header.ExtentSpace().Value > 0 {
		space_tree_id := int64(root_header.SpaceTreePageNumber())
		if space_tree_id > 0xff000000 {
			return errors.New(
				fmt.Sprintf("ID %v: Unsupported extent tree", id))
		}

		if space_tree_id > 0 {
			// Left node
			err := SpaceNodeVisitor(ctx, space_tree_id)
			if err != nil {
				return err
			}

			// Right node
			err = SpaceNodeVisitor(ctx, space_tree_id+1)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func SpaceNodeVisitor(ctx *ESEContext, id int64) error {
	fmt.Printf("SpaceNodeVisitor: %v\n", id)

	page := ctx.GetPage(id)
	if !page.Flags().IsSet("Root") ||
		!page.Flags().IsSet("SpaceTree") {
		return errors.New(fmt.Sprintf("ID %v: Not SpaceTree node", id))
	}

	if page.NextPage() != 0 {
		return errors.New(fmt.Sprintf("ID %v: Next Page not 0", id))
	}

	if page.PreviousPage() != 0 {
		return errors.New(fmt.Sprintf("ID %v: Prev Page not 0", id))
	}

	values := GetPageValues(ctx, page)
	if len(values) == 0 {
		return nil
	}

	if len(values) == 0 {
		return nil
	}

	for _, value := range values[1:] {
		if page.Flags().IsSet("Leaf") {
			reader := value.Reader()
			key_size := ParseUint16(reader, 0)
			child_page_id := ParseUint32(reader, 2+int64(key_size))
			fmt.Printf("child_page_id %v\n", child_page_id)
		}
	}

	return nil
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
		ctx.Profile.ESENT_LEAF_HEADER(
			&BufferReaderAt{values[0].Buffer}, 0).Dump()
	}

	for _, value := range values[1:] {
		if header.IsBranch() {
			ctx.Profile.ESENT_BRANCH_ENTRY(value.Reader(), 0).Dump()
		} else if header.IsLeaf() {
			if flags.IsSet("SpaceTree") {
				ctx.Profile.ESENT_SPACE_TREE_ENTRY(value.Reader(), 0).Dump()
			} else if flags.IsSet("Index") {
				ctx.Profile.ESENT_INDEX_ENTRY(value.Reader(), 0).Dump()
			} else if flags.IsSet("Long") {
				// TODO
			} else {
				ctx.Profile.ESENT_LEAF_ENTRY(value.Reader(), 0).Dump()
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

func (self *ESENT_LEAF_ENTRY) Dump() {
	fmt.Println(self.DebugString())
}

func (self *ESENT_LEAF_ENTRY) EntryData(value *Value) int64 {
	// Tag includes Local Page Key - skip it and the common page key
	if value.Tag.Flags()&TAG_COMMON > 0 {
		return self.Offset + 4 +
			int64(self.LocalPageKeySize())
	}

	return self.Offset + int64(self.CommonPageKeySize())
}

func (self *ESENT_BRANCH_HEADER) Dump() {
	fmt.Println(self.DebugString())
}

func (self *ESENT_BRANCH_ENTRY) Dump() {
	fmt.Println(self.DebugString())

	fmt.Printf("  ChildPageNumber: %#x\n", self.ChildPageNumber())
}

func (self *ESENT_SPACE_TREE_HEADER) Dump() {
	fmt.Println(self.DebugString())
}

func (self *ESENT_LEAF_HEADER) Dump() {
	fmt.Println(self.DebugString())
}

func (self *ESENT_BRANCH_ENTRY) ChildPageNumber() int64 {
	return int64(ParseUint32(self.Reader, self.Offset+int64(self.LocalPageKeySize())+2))
}

func WalkPages(ctx *ESEContext,
	id int64,
	cb func(header *PageHeader, id int64, value *Value)) {
	header := ctx.GetPage(id)
	values := GetPageValues(ctx, header)
	for _, value := range values[1:] {
		if header.IsLeaf() {
			cb(header, id, value)
		} else if header.IsBranch() {
			// Walk the branch
			branch := ctx.Profile.ESENT_BRANCH_ENTRY(value.Reader(), 0)
			WalkPages(ctx, branch.ChildPageNumber(), cb)
		}
	}
}
