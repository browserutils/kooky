// Parser based on https://github.com/SecureAuthCorp/impacket.git

package parser

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/Velocidex/ordereddict"
	"github.com/davecgh/go-spew/spew"
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

// The tag contains a single row.
// 00000000  09 00 7f 80 00 00 00 00  00 00 01 06 7f 2d 00 01  |.............-..|
// 00000010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 08  |................|
// 00000020  00 00 00 b1 00 00 00 ff  ff ff ff ff ff ff 7f ff  |................|
// 00000030  ff ff ff ff ff ff 7f c0  00 01 04 00 01 3a 00 76  |.............:.v|
// 00000040  00 65 00 72 00 73 00 69  00 6f 00 6e 00 00 00     |.e.r.s.i.o.n...|

// The key consumes the first 11 bytes:
// struct ESENT_LEAF_ENTRY @ 0x0:
// CommonPageKeySize: 0x0
// LocalPageKeySize: 0x9

// Followed by a data definition header:
//struct ESENT_DATA_DEFINITION_HEADER @ 0xb:
//   LastFixedType: 0x6
//   LastVariableDataType: 0x7f
//   VariableSizeOffset: 0x2d

// Column IDs below LastFixedSize will be stored in the fixed size
// portion. Column id below LastVariableDataType will be stored in the
// variable data section and higher than LastVariableDataType will be
// tagged.

// The fixed section starts immediately after the
// ESENT_DATA_DEFINITION_HEADER (offset 0xb + 4 = 0xf)

// Then the following columns consume their types:
// Column EntryId Identifier 1 Type Long long
// Column MinimizedRDomainHash Identifier 2 Type Long long
// Column MinimizedRDomainLength Identifier 3 Type Unsigned long
// Column IncludeSubdomains Identifier 4 Type Unsigned long
// Column Expires Identifier 5 Type Long long
// Column LastTimeUsed Identifier 6 Type Long long

// In the above example we have no variable sized columns, so we go
// straight to the tagged values:

// Then the tagged values are consumed
// Column RDomain Identifier 256 Type Long Text

func (self *Table) tagToRecord(value *Value) *ordereddict.Dict {
	tag := NewESENT_LEAF_ENTRY(self.ctx, value)

	if Debug {
		fmt.Printf("Processing row in Tag @ %d %#x (%#x)",
			value.Tag.Offset, value.Tag.ValueOffset(self.ctx),
			value.Tag.ValueSize(self.ctx))
		spew.Dump(value.Buffer)
		tag.Dump()
	}

	result := ordereddict.NewDict()

	var taggedItems map[uint32][]byte

	dd_header := self.ctx.Profile.ESENT_DATA_DEFINITION_HEADER(tag.Reader, tag.EntryData())

	// Start to parse immediately after the dd_header
	offset := dd_header.Offset + int64(dd_header.Size())

	if Debug {
		fmt.Println(dd_header.DebugString())
	}

	prevItemLen := int64(0)
	variableSizeOffset := dd_header.Offset + int64(dd_header.VariableSizeOffset())
	variableDataBytesProcessed := int64(dd_header.LastVariableDataType()-127) * 2

	// Iterate over the column definitions and decode each
	// identifier according to where it comes from.
	for _, column := range self.Columns.Keys() {
		catalog_any, _ := self.Columns.Get(column)
		catalog := catalog_any.(*ESENT_CATALOG_DATA_DEFINITION_ENTRY)
		identifier := catalog.Identifier()
		column_type := catalog.Column().ColumnType().Name

		if Debug {
			fmt.Printf("Column %v Identifier %v Type %v\n", column, identifier,
				catalog.Column().ColumnType().Name)
		}

		// Column is stored in the fixed section.
		if identifier <= uint32(dd_header.LastFixedType()) {
			space_usage := catalog.Column().SpaceUsage()

			switch column_type {
			case "Boolean":
				if space_usage == 1 {
					result.Set(column, ParseUint8(tag.Reader, offset) > 0)
				}

			case "Signed byte":
				if space_usage == 1 {
					result.Set(column, ParseUint8(tag.Reader, offset))
				}

			case "Signed short":
				if space_usage == 2 {
					result.Set(column, ParseInt16(tag.Reader, offset))
				}

			case "Unsigned short":
				if space_usage == 2 {
					result.Set(column, ParseUint16(tag.Reader, offset))
				}

			case "Signed long":
				if space_usage == 4 {
					result.Set(column, ParseInt32(tag.Reader, offset))
				}

			case "Unsigned long":
				if space_usage == 4 {
					result.Set(column, ParseUint32(tag.Reader, offset))
				}

			case "Single precision FP":
				if space_usage == 4 {
					result.Set(column, math.Float32frombits(
						ParseUint32(tag.Reader, offset)))
				}

			case "Double precision FP":
				if space_usage == 8 {
					result.Set(column, math.Float64frombits(
						ParseUint64(tag.Reader, offset)))
				}

			case "DateTime":
				if space_usage == 8 {
					// Some hair brained time serialization method
					// https://docs.microsoft.com/en-us/windows/win32/extensible-storage-engine/jet-coltyp
					days_since_1900 := math.Float64frombits(
						ParseUint64(tag.Reader, offset))
					// In python time.mktime((1900,1,1,0,0,0,0,365,0))
					result.Set(column, time.Unix(int64(days_since_1900*24*60*60)+
						-2208988800, 0))
				}

			case "Long long", "Currency":
				if space_usage == 8 {
					result.Set(column, ParseUint64(tag.Reader, offset))
				}

			default:
				fmt.Printf("Can not handle Column %v fixed data %v\n",
					column, catalog.Column().DebugString())
			}

			if Debug {
				fmt.Printf("Consumed %#x bytes of FIXED space from %#x\n",
					catalog.Column().SpaceUsage(), offset)
			}

			// Move our offset along
			offset += int64(catalog.Column().SpaceUsage())

			// Identifier is stored in the variable section
		} else if 127 < identifier &&
			identifier <= uint32(dd_header.LastVariableDataType()) {

			// Variable data type
			index := int64(identifier) - 127 - 1
			itemLen := int64(ParseUint16(tag.Reader, variableSizeOffset+index*2))

			if itemLen&0x8000 > 0 {
				// Empty Item
				itemLen = prevItemLen
				result.Set(column, nil)
			} else {
				switch column_type {
				case "Binary":
					result.Set(column, ParseString(tag.Reader,
						variableSizeOffset+variableDataBytesProcessed,
						itemLen-prevItemLen))

				case "Text":
					result.Set(column, ParseString(
						tag.Reader,
						variableSizeOffset+variableDataBytesProcessed,
						itemLen-prevItemLen))

				default:
					fmt.Printf("Can not handle Column %v variable data %v\n",
						column, catalog.Column().DebugString())
				}
			}

			if Debug {
				fmt.Printf("Consumed %#x bytes of VARIABLE space from %#x\n",
					itemLen-prevItemLen, variableDataBytesProcessed)
			}

			variableDataBytesProcessed += itemLen - prevItemLen
			prevItemLen = itemLen

			// Tagged values
		} else if identifier > 255 {
			if taggedItems == nil {
				taggedItems = ParseTaggedValues(
					self.ctx, getSlice(value.Buffer,
						uint64(variableDataBytesProcessed+
							variableSizeOffset),
						uint64(len(value.Buffer))))
			}

			buf, pres := taggedItems[identifier]
			if pres {
				switch column_type {
				case "Binary", "Long Binary":
					result.Set(column, hex.EncodeToString(buf))

				case "Long Text":
					result.Set(column, ParseTerminatedUTF16String(
						&BufferReaderAt{buf}, 0))

				default:
					if Debug {
						fmt.Printf("Can not handle Column %v tagged data %v\n",
							column, catalog.Column().DebugString())
					}
				}
			}
		}
	}

	return result
}

func (self *RecordTag) FlagSkip() uint64 {
	return 1
}

func getSlice(buffer []byte, start, end uint64) []byte {
	if end < start {
		return nil
	}

	length := uint64(len(buffer))

	if start < 0 {
		start = 0
	}

	if start > length {
		start = length
	}

	if end > length {
		end = length
	}

	return buffer[start:end]
}

func ParseTaggedValues(ctx *ESEContext, buffer []byte) map[uint32][]byte {
	result := make(map[uint32][]byte)

	if len(buffer) < 2 {
		return result
	}

	reader := &BufferReaderAt{buffer}
	first_record := ctx.Profile.RecordTag(reader, 0)
	prev_record := first_record

	// Iterate over all tag headers - the headers go until the
	// start of the first data segment
	for offset := uint64(first_record.Size()); offset < first_record.DataOffset(); offset += uint64(first_record.Size()) {
		record := ctx.Profile.RecordTag(reader, int64(offset))
		result[uint32(prev_record.Identifier())] = getSlice(buffer, prev_record.DataOffset()+
			prev_record.FlagSkip(), record.DataOffset())

		if Debug {
			fmt.Printf("Consumed %#x bytes of TAGGED space from %#x for tag %#x\n",
				record.DataOffset()-prev_record.DataOffset()-prev_record.FlagSkip(),
				prev_record.DataOffset()+prev_record.FlagSkip(),
				prev_record.Identifier())
		}

		prev_record = record
	}

	// Last record goes to the end of the buffer.
	result[uint32(prev_record.Identifier())] = getSlice(buffer, prev_record.DataOffset()+
		prev_record.FlagSkip(), uint64(len(buffer)))

	if Debug {
		fmt.Printf("Consumed %#x bytes of TAGGED space from %#x for tag %#x\n",
			uint64(len(buffer))-prev_record.DataOffset()-prev_record.FlagSkip(),
			prev_record.DataOffset()+prev_record.FlagSkip(),
			prev_record.Identifier())
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
			// Each tag stores a single row - all the
			// columns in the row are encoded in this tag.
			return cb(table.tagToRecord(value))
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

	// Catalog follows the dd header
	catalog := self.ctx.Profile.ESENT_CATALOG_DATA_DEFINITION_ENTRY(dd_header.Reader,
		dd_header.Offset+int64(dd_header.Size()))

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
		if self.currentTable == nil {
			return errors.New("Internal Error: No existing table when adding column")
		}
		self.currentTable.Columns.Set(itemName, catalog)

	case "CATALOG_TYPE_INDEX":
		if self.currentTable == nil {
			return errors.New("Internal Error: No existing table when adding index")
		}

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
		result += fmt.Sprintf("[%v] (FDP %#x):\n%sColumns\n", table.Name,
			table.FatherDataPageNumber, space)
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
