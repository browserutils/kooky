package utils

import (
	"fmt"
	"reflect"

	"github.com/go-sqlite/sqlite3"
)

type TableRow struct {
	columns map[string]int
	record  *sqlite3.Record
}

func (row TableRow) String(columnName string) (string, error) { return Value[string](row, columnName) }
func (row TableRow) BytesOrFallback(columnName string, fallback []byte) ([]byte, error) {
	return ValueOrFallback(row, columnName, fallback, false)
}
func (row TableRow) BytesStringOrFallback(columnName string, fallback []byte) ([]byte, error) {
	return ValueOrFallback(row, columnName, fallback, true)
}

func (row TableRow) Bool(columnName string) (bool, error) {
	rawValue, err := row.Value(columnName)
	if err != nil {
		return false, err
	}
	return rawValue != 0, nil
}

func (row TableRow) Int64(columnName string) (int64, error) {
	rawValue, err := row.Value(columnName)
	if err != nil {
		return 0, err
	}
	switch value := rawValue.(type) {
	case int64:
		return value, nil
	case uint64:
		if int64(value) < 0 {
			return 0, fmt.Errorf("expected column [%s] to be int64; got uint64 value that can't fit in int64: %d", columnName, rawValue)
		}
		return int64(value), nil
	case int32:
		return int64(value), nil
	case int:
		return int64(value), nil
	default:
		return 0, fmt.Errorf("expected column [%s] to be int64; got %T with value %[2]v", columnName, rawValue)
	}
}

func (row TableRow) Value(columnName string) (any, error) {
	if index, ok := row.columns[columnName]; !ok {
		return nil, fmt.Errorf("table doesn't have a column named [%s]", columnName)
	} else if count := len(row.columns); count <= index {
		return nil, fmt.Errorf("column named [%s] has index %d but row only has %d values", columnName, index, count)
	} else {
		return row.record.Values[index], nil
	}
}

func (row TableRow) ValueOrFallback(columnName string, fallback any) any {
	if index, ok := row.columns[columnName]; ok && index < len(row.columns) {
		return row.record.Values[index]
	}
	return fallback
}

func Value[T any](row TableRow, columnName string) (T, error) {
	var zero T
	if index, ok := row.columns[columnName]; !ok {
		return zero, fmt.Errorf("table doesn't have a column named [%s]", columnName)
	} else if count := len(row.columns); count <= index {
		return zero, fmt.Errorf("column named [%s] has index %d but row only has %d values", columnName, index, count)
	} else if v, ok := row.record.Values[index].(T); !ok {
		return zero, fmt.Errorf("expected column [%s] to be type %T; got type %[3]T with value %[3]v", columnName, zero, row.record.Values[index])
	} else {
		return v, nil
	}
}

func ValueOrFallback[T any](row TableRow, columnName string, fallback T, tryConvert bool) (T, error) {
	index, ok := row.columns[columnName]
	if !ok || index >= len(row.columns) || index < 0 {
		return fallback, fmt.Errorf("expected column [%s] does not exist", columnName)
	}
	v := row.record.Values[index]
	vt, ok := v.(T)
	if ok {
		return vt, nil
	}
	var zero T
	if !tryConvert {
		var zero T
		return fallback, fmt.Errorf("expected column [%s] to be type %T; got type %[3]T with value %[3]v", columnName, zero, v)
	}
	rv := reflect.ValueOf(v)
	rt := reflect.TypeFor[T]()
	if !rv.CanConvert(rt) {
		return fallback, fmt.Errorf("expected column [%s] to be type %T; got type %[3]T with value %[3]v; using fallback: %v", columnName, zero, v, fallback)
	}
	return rv.Convert(rt).Interface().(T), nil
}
