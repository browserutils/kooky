package kooky

import (
	"context"
	"testing"
)

var nidFilters = []Filter{
	Domain(`.google.com`),
	Name(`NID`),
}

func BenchmarkFirstMatch(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		cookie := TraverseCookies(ctx).FirstMatch(ctx, nidFilters...)
		_ = cookie
	}
}

func BenchmarkFirstMatchSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// cookies := ReadCookies(nidFilters...) // old name
		cookies := AllCookies(nidFilters...)
		_ = cookies
	}
}
