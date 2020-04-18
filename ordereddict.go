package ordereddict

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// A concerete implementation of a row - similar to Python's
// OrderedDict.  Main difference is that delete is not implemented -
// we just preserve the order of insertions.
type Dict struct {
	sync.Mutex

	store    map[string]interface{}
	keys     []string
	case_map map[string]string

	default_value interface{}
}

func NewDict() *Dict {
	return &Dict{
		store: make(map[string]interface{}),
	}
}

func (self *Dict) IsCaseInsensitive() bool {
	self.Lock()
	defer self.Unlock()

	return self.case_map != nil
}

func (self *Dict) MergeFrom(other *Dict) {
	for _, key := range other.keys {
		value, pres := other.Get(key)
		if pres {
			self.Set(key, value)
		}
	}
}

func (self *Dict) SetDefault(value interface{}) *Dict {
	self.Lock()
	defer self.Unlock()

	self.default_value = value
	return self
}

func (self *Dict) GetDefault() interface{} {
	self.Lock()
	defer self.Unlock()

	return self.default_value
}

func (self *Dict) SetCaseInsensitive() *Dict {
	self.Lock()
	defer self.Unlock()

	self.case_map = make(map[string]string)
	return self
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func (self *Dict) Set(key string, value interface{}) *Dict {
	self.Lock()
	defer self.Unlock()

	// O(n) but for our use case this is faster since Dicts are
	// typically small and we rarely overwrite a key.
	_, pres := self.store[key]
	if pres {
		self.keys = append(remove(self.keys, key), key)
	} else {
		self.keys = append(self.keys, key)
	}

	self.store[key] = value

	if self.case_map != nil {
		self.case_map[strings.ToLower(key)] = key
	}

	return self
}

func (self *Dict) Len() int {
	self.Lock()
	defer self.Unlock()

	return len(self.store)
}

func (self *Dict) Get(key string) (interface{}, bool) {
	self.Lock()
	defer self.Unlock()

	if self.case_map != nil {
		real_key, pres := self.case_map[strings.ToLower(key)]
		if pres {
			key = real_key
		}
	}

	val, ok := self.store[key]
	if !ok && self.default_value != nil {
		return self.default_value, false
	}

	return val, ok
}

func (self *Dict) Keys() []string {
	self.Lock()
	defer self.Unlock()

	return self.keys[:]
}

func (self *Dict) ToDict() *map[string]interface{} {
	self.Lock()
	defer self.Unlock()

	result := make(map[string]interface{})

	for _, key := range self.keys {
		value, pres := self.store[key]
		if pres {
			result[key] = value
		}
	}

	return &result
}

// Printing the dict will always result in a valid JSON document.
func (self *Dict) String() string {
	serialized, err := self.MarshalJSON()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return string(serialized)
}

func (self *Dict) GoString() string {
	return self.String()
}

// this implements type json.Unmarshaler interface, so can be called in json.Unmarshal(data, om)
func (self *Dict) UnmarshalJSON(data []byte) error {
	self.Lock()
	defer self.Unlock()

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	// must open with a delim token '{'
	t, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok || delim != '{' {
		return fmt.Errorf("expect JSON object open with '{'")
	}

	err = self.parseobject(dec)
	if err != nil {
		return err
	}

	t, err = dec.Token()
	if err != io.EOF {
		return fmt.Errorf("expect end of JSON object but got more token: %T: %v or err: %v", t, t, err)
	}

	return nil
}

func (self *Dict) parseobject(dec *json.Decoder) (err error) {
	var t json.Token
	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return err
		}

		key, ok := t.(string)
		if !ok {
			return fmt.Errorf("expecting JSON key should be always a string: %T: %v", t, t)
		}

		t, err = dec.Token()
		if err == io.EOF {
			break

		} else if err != nil {
			return err
		}

		var value interface{}
		value, err = handledelim(t, dec)
		if err != nil {
			return err
		}
		self.keys = append(self.keys, key)
		self.store[key] = value
		if self.case_map != nil {
			self.case_map[strings.ToLower(key)] = key
		}
	}

	t, err = dec.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok || delim != '}' {
		return fmt.Errorf("expect JSON object close with '}'")
	}

	return nil
}

func parsearray(dec *json.Decoder) (arr []interface{}, err error) {
	var t json.Token
	arr = make([]interface{}, 0)
	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return
		}

		var value interface{}
		value, err = handledelim(t, dec)
		if err != nil {
			return
		}
		arr = append(arr, value)
	}
	t, err = dec.Token()
	if err != nil {
		return
	}
	delim, ok := t.(json.Delim)

	if !ok || delim != ']' {
		err = fmt.Errorf("expect JSON array close with ']'")
		return
	}

	return
}

func handledelim(token json.Token, dec *json.Decoder) (res interface{}, err error) {
	switch t := token.(type) {
	case json.Delim:
		switch t {
		case '{':
			dict2 := NewDict()
			err = dict2.parseobject(dec)
			if err != nil {
				return
			}
			return dict2, nil
		case '[':
			var value []interface{}
			value, err = parsearray(dec)
			if err != nil {
				return
			}
			return value, nil
		default:
			return nil, fmt.Errorf("Unexpected delimiter: %q", t)
		}

	case json.Number:
		value, err := t.Int64()
		if err == nil {
			return value, nil
		}

		float, err := t.Float64()
		if err == nil {
			return float, nil
		}

		return nil, fmt.Errorf("Unexpected token: %v", token)
	}
	return token, nil
}

func (self Dict) MarshalJSON() ([]byte, error) {
	self.Lock()
	defer self.Unlock()

	result := "{"
	for _, k := range self.keys {

		// add key
		kEscaped, err := json.Marshal(k)
		if err != nil {
			continue
		}

		result += string(kEscaped) + ":"

		// add value
		v := self.store[k]

		vBytes, err := marshal(v)
		if err == nil {
			result += string(vBytes) + ","
		} else {
			result += "null,"
		}
	}
	if len(self.keys) > 0 {
		result = result[0 : len(result)-1]
	}
	result = result + "}"
	return []byte(result), nil
}

func marshal(v interface{}) ([]byte, error) {
	switch t := v.(type) {
	case time.Time:
		// Always marshal times as UTC
		return json.Marshal(t.UTC())

	case *time.Time:
		// Always marshal times as UTC
		return json.Marshal(t.UTC())

	default:
		return json.Marshal(v)
	}
}
