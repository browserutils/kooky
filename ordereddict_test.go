package ordereddict

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/alecthomas/repr"
	"github.com/stretchr/testify/assert"
)

type dictSerializationTest struct {
	dict       *Dict
	serialized string
}

var (
	dictSerializationTests = []dictSerializationTest{
		{NewDict().Set("Foo", "Bar"), `{"Foo":"Bar"}`},

		// Test an unserilizable member - This should not prevent the
		// entire dict from serializing - only that member should be
		// ignored.
		{NewDict().Set("Foo", "Bar").
			Set("Time", time.Unix(3000000000000000, 0)),
			`{"Foo":"Bar","Time":null}`},

		// Recursive dict
		{NewDict().Set("Foo",
			NewDict().Set("Bar", 2).
				Set("Time", time.Unix(3000000000000000, 0))),
			`{"Foo":{"Bar":2,"Time":null}}`},

		// Ensure key order is preserved.
		{NewDict().Set("A", 1).Set("B", 2), `{"A":1,"B":2}`},
		{NewDict().Set("B", 1).Set("A", 2), `{"B":1,"A":2}`},

		// Serialize with quotes
		{NewDict().Set("foo\\'s quote", 1), `{"foo\\'s quote":1}`},
	}

	// Check that serialization decodes to the object.
	dictUnserializationTest = []dictSerializationTest{
		// Preserve order of keys on deserialization.
		{NewDict().Set("A", uint64(1)).Set("B", uint64(2)), `{"A":1,"B":2}`},
		{NewDict().Set("B", uint64(1)).Set("A", uint64(2)), `{"B":1,"A":2}`},

		// Handle arrays, ints floats and bools
		{NewDict().Set("B", uint64(1)).Set("A", uint64(2)), `{"B":1,"A":2}`},
		{NewDict().Set("B", float64(1)).Set("A", uint64(2)), `{"B":1.0,"A":2}`},
		{NewDict().Set("B", []interface{}{uint64(1)}).Set("A", uint64(2)), `{"B":[1],"A":2}`},
		{NewDict().Set("B", true).Set("A", uint64(2)), `{"B":true,"A":2}`},
		{NewDict().Set("B", nil).Set("A", uint64(2)), `{"B":null,"A":2}`},

		// Embedded dicts decode into ordered dicts.
		{NewDict().
			Set("B", NewDict().Set("Zoo", "X").Set("Baz", "Y")).
			Set("A", "Z"), `{"B":{"Zoo":"X","Baz":"Y"},"A":"Z"}`},

		// Make sure we properly preserve uint64 (overflows int64)
		{NewDict().
			Set("Uint64", uint64(9223372036854775808)),
			`{"Uint64": 9223372036854775808}`},

		// We prefer uint64 but int64 is needed for negative numbers
		{NewDict().
			Set("Int64", int64(-500)),
			`{"Int64": -500}`},
	}
)

func TestDictSerialization(t *testing.T) {
	for _, test := range dictSerializationTests {
		serialized, err := json.Marshal(test.dict)
		if err != nil {
			t.Fatalf("Failed to serialize %v: %v", repr.String(test.dict), err)
		}

		assert.Equal(t, test.serialized, string(serialized))
	}
}

func TestDictDeserialization(t *testing.T) {
	for _, test := range dictUnserializationTest {
		value := NewDict()

		err := json.Unmarshal([]byte(test.serialized), value)
		if err != nil {
			t.Fatalf("Failed to serialize %v: %v", test.serialized, err)
		}
		// Make sure the keys are the same.
		assert.Equal(t, test.dict.Keys(), value.Keys())
		assert.Equal(t, test.dict.ToDict(), value.ToDict())

		assert.True(t, reflect.DeepEqual(test.dict, value))
	}
}

func TestOrder(t *testing.T) {
	test := NewDict().
		Set("A", 1).
		Set("B", 2)

	assert.Equal(t, []string{"A", "B"}, test.Keys())

	test = NewDict().
		Set("B", 1).
		Set("A", 2)

	assert.Equal(t, []string{"B", "A"}, test.Keys())
}

func TestCaseInsensitive(t *testing.T) {
	test := NewDict().SetCaseInsensitive()

	test.Set("FOO", 1)

	value, pres := test.Get("foo")
	assert.True(t, pres)
	assert.Equal(t, 1, value)

	test = NewDict().Set("FOO", 1)
	value, pres = test.Get("foo")
	assert.False(t, pres)
}
