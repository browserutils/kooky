package parser

import (
	"errors"
	"fmt"

	"github.com/Velocidex/ordereddict"
	"github.com/davecgh/go-spew/spew"
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
	leaf_entry := self.ctx.Profile.ESENT_LEAF_ENTRY(value.Reader(), 0)
	dd_header := self.ctx.Profile.ESENT_DATA_DEFINITION_HEADER(
		leaf_entry.Reader, leaf_entry.EntryData(value))

	itemName := parseItemName(dd_header)
	catalog := dd_header.Catalog()

	switch catalog.Type().Name {
	case "CATALOG_TYPE_TABLE":
		table := &Table{
			Header:               catalog.Table(),
			Name:                 itemName,
			FatherDataPageNumber: catalog.FDPId(),
			Columns:              ordereddict.NewDict(),
			Indexes:              ordereddict.NewDict(),
			LongValues:           ordereddict.NewDict()}
		self.currentTable = table
		self.Tables.Set(itemName, table)

	case "CATALOG_TYPE_COLUMN":
		self.currentTable.Columns.Set(itemName, catalog.Column())

	case "CATALOG_TYPE_INDEX":
		self.currentTable.Indexes.Set(itemName, catalog.Index())
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
	leaf_entry := self.ctx.Profile.ESENT_LEAF_ENTRY(value.Reader(), 0)
	fmt.Printf("Leaf %v\n", leaf_entry.DebugString())
	spew.Dump(value.Buffer)
}

func (self *Catalog) Dump() {
	for _, name := range self.Tables.Keys() {
		table_any, _ := self.Tables.Get(name)
		table := table_any.(*Table)

		space := "   "
		fmt.Printf("[%v]:\n%sColumns\n", table.Name, space)
		for idx, column := range table.Columns.Keys() {
			column_header_any, _ := table.Columns.Get(column)
			column_header := column_header_any.(*CATALOG_TYPE_COLUMN)
			fmt.Printf("%s%s%-5d%-30v%v\n", space, space, idx,
				column, column_header.ColumnType().Name)
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
