package ordereddict

import (
	"reflect"
	"strings"
	"sync"
)

// A concrete implementation of a row - similar to Python's
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

	if self.store == nil {
		self.store = make(map[string]interface{})
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
