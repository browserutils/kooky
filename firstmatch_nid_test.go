package kooky_test

import (
	"context"
	"testing"

	"github.com/browserutils/kooky"
	"github.com/browserutils/kooky/filter"
)

var nidFilters = []kooky.Filter{
	filter.Domain(`.google.com`),
	filter.Name(`NID`),
}

func BenchmarkFirstMatch(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		cookie := kooky.TraverseCookies(ctx).FirstMatch(ctx, nidFilters...)
		_ = cookie
	}
}

func BenchmarkFirstMatchSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// cookies := ReadCookies(nidFilters...) // old name
		cookies := kooky.AllCookies(nidFilters...)
		_ = cookies
	}
}
