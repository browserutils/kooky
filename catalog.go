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
	ctx                  *ESEContext
	Header               *CATALOG_TYPE_TABLE
	FatherDataPageNumber uint32
	Name                 string
	Columns              *ordereddict.Dict
	Indexes              *ordereddict.Dict
	LongValues           *ordereddict.Dict
}

// Code based on
func (self *Table) tagToRecord(tag *ESENT_LEAF_ENTRY) *ordereddict.Dict {
	result := ordereddict.NewDict()

	dd_header := self.ctx.Profile.ESENT_DATA_DEFINITION_HEADER(tag.Reader, tag.EntryData())
	fixed_size_offset := dd_header.Offset + self.ctx.Profile.
		Off_ESENT_DATA_DEFINITION_HEADER_FixedSizeStart
	prevItemLen := int64(0)
	variableSizeOffset := dd_header.Offset + int64(dd_header.VariableSizeOffset())
	variableDataBytesProcessed := int64(dd_header.LastVariableDataType()-127) * 2

	for _, column := range self.Columns.Keys() {
		catalog_any, _ := self.Columns.Get(column)
		catalog := catalog_any.(*ESENT_CATALOG_DATA_DEFINITION_ENTRY)

		if catalog.Identifier() <= uint32(dd_header.LastFixedSize()) {
			space_usage := catalog.Column().SpaceUsage()

			switch catalog.Column().ColumnType().Name {
			case "Boolean":
				if space_usage == 1 {
					result.Set(column, ParseUint8(tag.Reader, fixed_size_offset) > 0)
				}

			case "Signed byte":
				if space_usage == 1 {
					result.Set(column, ParseUint8(tag.Reader, fixed_size_offset))
				}

			case "Signed short":
				if space_usage == 2 {
					result.Set(column, ParseInt16(tag.Reader, fixed_size_offset))
				}

			case "Unsigned short":
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

// DumpTable extracts all rows in the named table and passes them into
// the callback. The callback may cancel the walk at any time by
// returning an error which is passed to our caller.
func (self *Catalog) DumpTable(name string, cb func(row *ordereddict.Dict) error) error {
	table_any, pres := self.Tables.Get(name)
	if !pres {
		return errors.New("Table not found")
	}

	table := table_any.(*Table)
	err := WalkPages(self.ctx, int64(table.FatherDataPageNumber),
		func(header *PageHeader, id int64, value *Value) error {
			leaf_entry := NewESENT_LEAF_ENTRY(self.ctx, value)
			return cb(table.tagToRecord(leaf_entry))
		})
	if err != nil {
		return err
	}
	return nil
}

// Catalog represents the database's catalog.
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

func (self *Catalog) __addItem(header *PageHeader, id int64, value *Value) error {
	leaf_entry := NewESENT_LEAF_ENTRY(self.ctx, value)
	dd_header := self.ctx.Profile.ESENT_DATA_DEFINITION_HEADER(
		leaf_entry.Reader, leaf_entry.EntryData())

	itemName := parseItemName(dd_header)
	catalog := dd_header.Catalog()

	switch catalog.Type().Name {
	case "CATALOG_TYPE_TABLE":
		table := &Table{
			ctx:                  self.ctx,
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

	return nil
}

func (self *Catalog) Dump() string {
	result := ""

	for _, name := range self.Tables.Keys() {
		table_any, _ := self.Tables.Get(name)
		table := table_any.(*Table)

		space := "   "
		result += fmt.Sprintf("[%v]:\n%sColumns\n", table.Name, space)
		for idx, column := range table.Columns.Keys() {
			catalog_any, _ := table.Columns.Get(column)
			catalog := catalog_any.(*ESENT_CATALOG_DATA_DEFINITION_ENTRY)
			result += fmt.Sprintf("%s%s%-5d%-30v%v\n", space, space, idx,
				column, catalog.Column().ColumnType().Name)
		}

		result += fmt.Sprintf("%sIndexes\n", space)
		for _, index := range table.Indexes.Keys() {
			result += fmt.Sprintf("%s%s%v:\n", space, space, index)
		}
		result += "\n"
	}

	return result
}

func ReadCatalog(ctx *ESEContext) (*Catalog, error) {
	result := &Catalog{ctx: ctx, Tables: ordereddict.NewDict()}

	err := WalkPages(ctx, CATALOG_PAGE_NUMBER, result.__addItem)
	if err != nil {
		return nil, err
	}
	return result, nil
}
