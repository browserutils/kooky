// Parser based on https://github.com/SecureAuthCorp/impacket.git

package parser

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/Velocidex/ordereddict"
)

const (
	CATALOG_PAGE_NUMBER = 4
)

type Table struct {
	Header               *CATALOG_TYPE_TABLE
	FatherDataPageNumber uint32
	Name                 string
	Columns              *ordereddict.Dict
	Indexes              *ordereddict.Dict
	LongValues           *ordereddict.Dict
}

type Catalog struct {
	ctx *ESEContext

	Tables *ordereddict.Dict

	currentTable *Table
}

func parseItemName(dd_header *ESENT_DATA_DEFINITION_HEADER) string {
	last_variable_data_type := int64(dd_header.LastVariableDataType())
	numEntries := last_variable_data_type

	if last_variable_data_type > 127 {
		numEntries = last_variable_data_type - 127
	}

	itemLen := ParseUint16(dd_header.Reader,
		dd_header.Offset+int64(dd_header.VariableSizeOffset()))

	return ParseString(dd_header.Reader,
		dd_header.Offset+int64(dd_header.VariableSizeOffset())+
			2*numEntries, int64(itemLen))
}

func (self *Catalog) __addItem(header *PageHeader, id int64, value *Value) {
	leaf_entry := NewESENT_LEAF_ENTRY(self.ctx, value)
	dd_header := self.ctx.Profile.ESENT_DATA_DEFINITION_HEADER(
		leaf_entry.Reader, leaf_entry.EntryData())

	itemName := parseItemName(dd_header)
	catalog := dd_header.Catalog()

	switch catalog.Type().Name {
	case "CATALOG_TYPE_TABLE":
		table := &Table{
			Header:               catalog.Table(),
			Name:                 itemName,
			FatherDataPageNumber: catalog.Table().FatherDataPageNumber(),
			Columns:              ordereddict.NewDict(),
			Indexes:              ordereddict.NewDict(),
			LongValues:           ordereddict.NewDict()}
		self.currentTable = table
		self.Tables.Set(itemName, table)

	case "CATALOG_TYPE_COLUMN":
		self.currentTable.Columns.Set(itemName, catalog)

	case "CATALOG_TYPE_INDEX":
		self.currentTable.Indexes.Set(itemName, catalog)
	case "CATALOG_TYPE_LONG_VALUE":

	}
}

func (self *Catalog) DumpTable(name string) error {
	table_any, pres := self.Tables.Get(name)
	if !pres {
		return errors.New("Table not found")
	}

	table := table_any.(*Table)
	WalkPages(self.ctx, int64(table.FatherDataPageNumber), self._walkTableContents)

	return nil
}

func (self *Catalog) _walkTableContents(header *PageHeader, id int64, value *Value) {
	leaf_entry := NewESENT_LEAF_ENTRY(self.ctx, value)
	fmt.Printf("Leaf @ %v \n", id)
	_ = leaf_entry
	//	spew.Dump(value.Buffer)
}

func (self *Catalog) Dump() {
	for _, name := range self.Tables.Keys() {
		table_any, _ := self.Tables.Get(name)
		table := table_any.(*Table)

		space := "   "
		fmt.Printf("[%v]:\n%sColumns\n", table.Name, space)
		for idx, column := range table.Columns.Keys() {
			catalog_any, _ := table.Columns.Get(column)
			catalog := catalog_any.(*ESENT_CATALOG_DATA_DEFINITION_ENTRY)
			fmt.Printf("%s%s%-5d%-30v%v\n", space, space, idx,
				column, catalog.Column().ColumnType().Name)
		}

		fmt.Printf("%sIndexes\n", space)
		for _, index := range table.Indexes.Keys() {
			fmt.Printf("%s%s%v:\n", space, space, index)
		}
		fmt.Printf("\n")
	}
}

func ReadCatalog(ctx *ESEContext) *Catalog {
	result := &Catalog{ctx: ctx, Tables: ordereddict.NewDict()}

	WalkPages(ctx, CATALOG_PAGE_NUMBER, result.__addItem)
	return result
}

type Cursor struct {
	ctx                  *ESEContext
	FatherDataPageNumber uint64
	Table                *Table
	CurrentPageData      *PageHeader
	CurrentTag           int
	CurrentValues        []*Value
}

func (self *Catalog) OpenTable(ctx *ESEContext, name string) (*Cursor, error) {
	table_any, pres := self.Tables.Get(name)
	if !pres {
		return nil, errors.New("Table not found")
	}

	table := table_any.(*Table)
	catalog_entry := table.Header
	pageNum := int64(catalog_entry.FatherDataPageNumber())
	done := false
	var page *PageHeader
	var values []*Value

	for !done {
		page = ctx.GetPage(pageNum)
		values = GetPageValues(ctx, page)

		if len(values) > 1 {
			for _, value := range values[1:] {
				if page.IsBranch() {
					branchEntry := NewESENT_BRANCH_ENTRY(ctx, value)
					pageNum = branchEntry.ChildPageNumber()
					break
				} else {
					done = true
					break
				}
			}
		}
	}

	return &Cursor{
		ctx:                  ctx,
		Table:                table,
		FatherDataPageNumber: uint64(catalog_entry.FatherDataPageNumber()),
		CurrentPageData:      page,
		CurrentTag:           0,
		CurrentValues:        values,
	}, nil
}

func (self *Cursor) GetNextTag() *ESENT_LEAF_ENTRY {
	page := self.CurrentPageData

	if self.CurrentTag >= len(self.CurrentValues) {
		return nil
	}

	if page.IsBranch() {
		return nil
	}

	page_flags := page.Flags()
	if page_flags.IsSet("SpaceTree") ||
		page_flags.IsSet("Index") ||
		page_flags.IsSet("Long") {

		// Log this exception.
		return nil
	}

	return NewESENT_LEAF_ENTRY(self.ctx, self.CurrentValues[self.CurrentTag])
}

func (self *Cursor) GetNextRow() *ordereddict.Dict {
	self.CurrentTag++

	tag := self.GetNextTag()
	if tag == nil {
		page := self.CurrentPageData
		if page.NextPageNumber() == 0 {
			return nil
		}

		self.CurrentPageData = self.ctx.GetPage(int64(page.NextPageNumber()))
		self.CurrentTag = 0
		self.CurrentValues = GetPageValues(self.ctx, self.CurrentPageData)
		return self.GetNextRow()
	}

	return self.tagToRecord(tag)
}

func (self *Cursor) tagToRecord(tag *ESENT_LEAF_ENTRY) *ordereddict.Dict {
	result := ordereddict.NewDict()

	dd_header := self.ctx.Profile.ESENT_DATA_DEFINITION_HEADER(tag.Reader, tag.EntryData())
	fixed_size_offset := dd_header.Offset + self.ctx.Profile.
		Off_ESENT_DATA_DEFINITION_HEADER_FixedSizeStart
	prevItemLen := int64(0)
	variableSizeOffset := dd_header.Offset + int64(dd_header.VariableSizeOffset())
	variableDataBytesProcessed := int64(dd_header.LastVariableDataType()-127) * 2

	for _, column := range self.Table.Columns.Keys() {
		catalog_any, _ := self.Table.Columns.Get(column)
		catalog := catalog_any.(*ESENT_CATALOG_DATA_DEFINITION_ENTRY)

		if catalog.Identifier() <= uint32(dd_header.LastFixedSize()) {
			space_usage := catalog.Column().SpaceUsage()

			switch catalog.Column().ColumnType().Name {
			case "Signed byte":
				if space_usage == 1 {
					result.Set(column, ParseUint8(tag.Reader, fixed_size_offset))
				}
			case "Signed short":
				if space_usage == 2 {
					result.Set(column, ParseUint16(tag.Reader, fixed_size_offset))
				}

			case "Signed long":
				if space_usage == 4 {
					result.Set(column, ParseInt32(tag.Reader, fixed_size_offset))
				}

			case "Unsigned long":
				if space_usage == 4 {
					result.Set(column, ParseUint32(tag.Reader, fixed_size_offset))
				}

			case "Single precision FP":
				if space_usage == 4 {
					result.Set(column, math.Float32frombits(
						ParseUint32(tag.Reader, fixed_size_offset)))
				}

			case "Double precision FP":
				if space_usage == 8 {
					result.Set(column, math.Float64frombits(
						ParseUint64(tag.Reader, fixed_size_offset)))
				}

			case "DateTime":
				if space_usage == 8 {
					// Some hair brained time serialization method
					// https://docs.microsoft.com/en-us/windows/win32/extensible-storage-engine/jet-coltyp
					days_since_1900 := math.Float64frombits(
						ParseUint64(tag.Reader, fixed_size_offset))
					// In python time.mktime((1900,1,1,0,0,0,0,365,0))
					result.Set(column, time.Unix(int64(days_since_1900*24*60*60)+
						-2208988800, 0))
				}

			case "Long long", "Currency":
				if space_usage == 8 {
					result.Set(column, ParseUint64(tag.Reader, fixed_size_offset))
				}

			default:
				fmt.Printf("Can not handle %v\n", catalog.Column().DebugString())
			}
			fixed_size_offset += int64(catalog.Column().SpaceUsage())

		} else if 127 < catalog.Identifier() &&
			catalog.Identifier() <= uint32(dd_header.LastVariableDataType()) {

			// Variable data type
			index := int64(catalog.Identifier()) - 127 - 1
			itemLen := int64(ParseUint16(tag.Reader, variableSizeOffset+index*2))

			if itemLen&0x8000 > 0 {
				// Empty Item
				itemLen = prevItemLen
				result.Set(column, nil)
			} else {
				data := ParseString(tag.Reader,
					variableSizeOffset+variableDataBytesProcessed,
					itemLen-prevItemLen)

				switch catalog.Column().ColumnType().Name {
				case "Text", "Binary":
					result.Set(column, data)
				}
			}

			variableDataBytesProcessed += itemLen - prevItemLen
			prevItemLen = itemLen
		}
	}

	return result
}
