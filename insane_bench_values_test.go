package insaneJSON

import (
	"testing"
)

func BenchmarkValueDecodeInt(b *testing.B) {
	tests := []struct {
		s string
		n int64
	}{
		{s: "", n: 0},
		{s: " ", n: 0},
		{s: "xxx", n: 0},
		{s: "-xxx", n: 0},
		{s: "1xxx", n: 0},
		{s: "-", n: 0},
		{s: "111 ", n: 0},
		{s: "1-1", n: 0},
		{s: "s1", n: 0},
		{s: "0", n: 0},
		{s: "-0", n: 0},
		{s: "5", n: 5},
		{s: "-5", n: -5},
		{s: " 0", n: 0},
		{s: " 5", n: 0},
		{s: "333", n: 333},
		{s: "-333", n: -333},
		{s: "1111111111", n: 1111111111},
		{s: "987654321", n: 987654321},
		{s: "123456789", n: 123456789},
		{s: "9223372036854775807", n: 9223372036854775807},
		{s: "-9223372036854775807", n: -9223372036854775807},
		{s: "9999999999999999999", n: 0},
		{s: "99999999999999999999", n: 0},
		{s: "-9999999999999999999", n: 0},
		{s: "-99999999999999999999", n: 0},
	}

	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			decodeInt64(test.s)
		}
	}

}

func BenchmarkValueEscapeString(b *testing.B) {
	tests := []struct {
		s string
	}{
		{s: `"""\\\\\"""\'\"				\\\""|"|"|"|\\'\dasd'		|"|\\\\'\\\|||\\'"`},
		{s: `sfsafwefqwueibfiquwbfiuqwebfiuqwbfiquwbfqiwbfoqiwuefb""""""""""""""""""""""""`},
		{s: `sfsafwefqwueibfiquwbfiuqwebfiuqwbfiquwbfqiwbfoqiwuefbxxxxxxxxxxxxxxxxxxxxxxx"`},
	}

	out := make([]byte, 0, 0)
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			out = escapeString(out[:0], test.s)
		}
	}
}
