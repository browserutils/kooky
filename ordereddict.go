package ordereddict

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/Velocidex/json"
	"github.com/Velocidex/yaml/v2"
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

// Very inefficient but ok for occasional use.
func (self *Dict) Delete(key string) {
	new_keys := make([]string, 0, len(self.keys))
	for _, old_key := range self.keys {
		if key != old_key {
			new_keys = append(new_keys, old_key)
		}
	}
	self.keys = new_keys
	delete(self.store, key)
}

// Like Set() but does not effect the order.
func (self *Dict) Update(key string, value interface{}) *Dict {
	self.Lock()
	defer self.Unlock()

	_, pres := self.store[key]
	if pres {
		self.store[key] = value
	} else {
		self.set(key, value)
	}

	return self
}

func (self *Dict) Set(key string, value interface{}) *Dict {
	self.Lock()
	defer self.Unlock()

	return self.set(key, value)
}

func (self *Dict) set(key string, value interface{}) *Dict {
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

func (self *Dict) GetString(key string) (string, bool) {
	v, pres := self.Get(key)
	if pres {
		v_str, ok := to_string(v)
		if ok {
			return v_str, true
		}
	}
	return "", false
}

func (self *Dict) GetBool(key string) (bool, bool) {
	v, pres := self.Get(key)
	if pres {
		v_bool, ok := v.(bool)
		if ok {
			return v_bool, true
		}
	}
	return false, false
}

func to_string(x interface{}) (string, bool) {
	switch t := x.(type) {
	case string:
		return t, true
	case *string:
		return *t, true
	case []byte:
		return string(t), true
	default:
		return "", false
	}
}

func (self *Dict) GetStrings(key string) ([]string, bool) {
	v, pres := self.Get(key)
	if pres && v != nil {
		slice := reflect.ValueOf(v)
		if slice.Type().Kind() == reflect.Slice {
			result := []string{}
			for i := 0; i < slice.Len(); i++ {
				value := slice.Index(i).Interface()
				item, ok := to_string(value)
				if ok {
					result = append(result, item)
				}
			}
			return result, true
		}
	}
	return nil, false
}

func (self *Dict) GetInt64(key string) (int64, bool) {
	value, pres := self.Get(key)
	if pres {
		switch t := value.(type) {
		case int:
			return int64(t), true
		case int8:
			return int64(t), true
		case int16:
			return int64(t), true
		case int32:
			return int64(t), true
		case int64:
			return int64(t), true
		case uint8:
			return int64(t), true
		case uint16:
			return int64(t), true
		case uint32:
			return int64(t), true
		case uint64:
			return int64(t), true
		case float32:
			return int64(t), true
		case float64:
			return int64(t), true
		}
	}
	return 0, false
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

// this implements type json.Unmarshaler interface, so can be called
// in json.Unmarshal(data, om). We preserve key order when
// unmarshaling from JSON.
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
		value_str := t.String()

		// Try to parse as Uint
		value_uint, err := strconv.ParseUint(value_str, 10, 64)
		if err == nil {
			return value_uint, nil
		}

		value_int, err := strconv.ParseInt(value_str, 10, 64)
		if err == nil {
			return value_int, nil
		}

		// Failing this, try a float
		float, err := strconv.ParseFloat(value_str, 64)
		if err == nil {
			return float, nil
		}

		return nil, fmt.Errorf("Unexpected token: %v", token)
	}
	return token, nil
}

// Preserve key order when marshalling to JSON.
func (self *Dict) MarshalJSON() ([]byte, error) {
	self.Lock()
	defer self.Unlock()

	buf := &bytes.Buffer{}
	buf.Write([]byte("{"))
	for _, k := range self.keys {

		// add key
		kEscaped, err := json.Marshal(k)
		if err != nil {
			continue
		}

		// add value
		v := self.store[k]

		// Check for back references and skip them - this is not perfect.
		subdict, ok := v.(*Dict)
		if ok && subdict == self {
			continue
		}

		buf.Write(kEscaped)
		buf.Write([]byte(":"))

		vBytes, err := json.Marshal(v)
		if err == nil {
			buf.Write(vBytes)
			buf.Write([]byte(","))
		} else {
			buf.Write([]byte("null,"))
		}
	}
	if len(self.keys) > 0 {
		buf.Truncate(buf.Len() - 1)
	}
	buf.Write([]byte("}"))
	return buf.Bytes(), nil
}

func (self *Dict) MarshalYAML() (interface{}, error) {
	self.Lock()
	defer self.Unlock()

	result := yaml.MapSlice{}
	for _, k := range self.keys {
		v := self.store[k]
		result = append(result, yaml.MapItem{
			Key: k, Value: v,
		})
	}

	return result, nil
}
