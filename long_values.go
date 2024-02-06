package parser

import (
	"fmt"
	"io"
)

type LongValue struct {
	Value  *Value
	header *PageHeader

	// Calculated key for this long value object.
	Key Key
}

func (self *LongValue) Buffer() []byte {
	start := int64(self.Key.EndOffset())
	result := make([]byte, self.Value.BufferSize-start)
	self.Value.Reader().ReadAt(result, start)
	return result
}

func (self *LongValue) Reader() io.ReaderAt {
	start := int64(self.Key.EndOffset())
	return NewOffsetReader(
		self.Value.Reader(), start, self.Value.BufferSize-start)
}

type LongValueLookup map[string]*LongValue

// This can potentially return a lot of data
func (self LongValueLookup) GetLid(lid []byte) ([]byte, bool) {
	if len(lid) != 4 {
		return nil, false
	}

	// Swap byte order between Lid and key
	swapped := []byte{lid[3], lid[2], lid[1], lid[0]}

	// For now we only get the first segment
	key := Key{
		prefix: swapped,
		suffix: make([]byte, 4),
	}

	// Try to find segments
	value, pres := self[key.Key()]
	if pres {
		return value.Buffer(), true
	}
	return nil, false
}

func NewLongValueLookup() LongValueLookup {
	return make(LongValueLookup)
}

type Key struct {
	prefix []byte
	suffix []byte

	end_offset uint64
}

func (self *Key) Key() string {
	return string(self.prefix) + string(self.suffix)
}

func (self *Key) DebugString() string {
	result := ""
	if len(self.prefix) > 0 {
		result += fmt.Sprintf("prefix %02x ", self.prefix)
	}

	if len(self.suffix) > 0 {
		result += fmt.Sprintf("suffix %02x ", self.suffix)
	}
	return result + fmt.Sprintf(" key %02x ", self.Key())
}

func (self *Key) EndOffset() uint64 {
	return self.end_offset
}

func (self *LVKEY_BUFFER) ParseKey(ctx *ESEContext, header *PageHeader, value *Value) (key Key) {
	key.end_offset = uint64(ctx.Profile.Off_LVKEY_BUFFER_KeyBuffer)

	prefix_lenth := uint64(self.PrefixLength())
	if prefix_lenth > 8 {
		prefix_lenth = 8
	}

	suffix_length := uint64(self.SuffixLength())
	if suffix_length > 8-prefix_lenth {
		suffix_length = 8 - prefix_lenth
	}

	// Compressed keys
	if value.Tag.Flags_()&fNDCompressed > 0 {
		external_value := header.ExternalValueBytes(ctx)
		if prefix_lenth > uint64(len(external_value)) {
			prefix_lenth = uint64(len(external_value))
		}

		for i := uint64(0); i < prefix_lenth; i++ {
			key.prefix = append(key.prefix, external_value[i])
		}
		key_buffer := self.KeyBuffer()
		if suffix_length > 8 {
			suffix_length = 8
		}

		key.suffix = []byte(key_buffer[:suffix_length])
		key.end_offset += suffix_length

	} else {
		// The key is not compressed - we read both prefix
		// and suffix from the actual key
		key_buffer := []byte(self.KeyBuffer())
		for uint64(len(key_buffer)) < prefix_lenth+suffix_length {
			key_buffer = append(key_buffer, 0)
		}

		key.prefix = key_buffer[:prefix_lenth]
		key.suffix = key_buffer[prefix_lenth : prefix_lenth+suffix_length]
		key.end_offset += suffix_length + prefix_lenth
	}

	return key
}
