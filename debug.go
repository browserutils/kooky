package parser

import "fmt"

var (
	// General purpose debug statements.
	Debug = false

	// Debugging during walk
	DebugWalk = false
)

func DlvDebug() {

}

func DebugPageHeader(ctx *ESEContext, page *PageHeader) string {
	return page.DebugString() + fmt.Sprintf("  EndOffset: %#x \n", page.EndOffset(ctx))
}

func DebugTag(ctx *ESEContext, tag *Tag, page *PageHeader) string {
	return tag.DebugString() +
		fmt.Sprintf("   ValueOffsetInPage: %#x \n",
			tag.ValueOffsetInPage(ctx, page))
}
