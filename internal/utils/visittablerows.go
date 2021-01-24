package utils

import (
	"fmt"

	"github.com/go-sqlite/sqlite3"
)

func VisitTableRows(db *sqlite3.DbFile, tableName string, f func(rowID *int64, row TableRow) error) error {
	columns := make(map[string]int)
	if table, ok := findTable(db, tableName); ok {
		for index, column := range table.Columns() {
			if _, ok := columns[column.Name()]; !ok {
				columns[column.Name()] = index
			}
		}
	} else {
		return fmt.Errorf("Unable to find table named [%s] in %v", tableName, db)
	}
	return db.VisitTableRecords(tableName, func(rowID *int64, record sqlite3.Record) error {
		return f(rowID, TableRow{columns, &record})
	})
}

func findTable(db *sqlite3.DbFile, tableName string) (sqlite3.Table, bool) {
	for _, table := range db.Tables() {
		if table.Name() == tableName {
			return table, true
		}
	}
	return sqlite3.Table{}, false
}
