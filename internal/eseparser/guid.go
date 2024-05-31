package parser

import "fmt"

func (self GUID) AsString() string {
	data4 := self.Data4()
	return fmt.Sprintf(
		"{%08x-%04x-%04x-%02x%02x-%02x%02x%02x%02x%02x%02x}", self.Data1(),
		self.Data2(), self.Data3(),
		data4[0], data4[1], data4[2], data4[3],
		data4[4], data4[5], data4[6], data4[7])
}
