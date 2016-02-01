package fser

import (
	"fmt"
	"testing"
)

func TestParsing(t *testing.T) {
	testCases := []struct {
		uri     string
		mustErr bool
	}{
		{
			uri:     "file:///Users/emmanuelodeke/go/src/github.com/odeke-em/drive/src/copy.go",
			mustErr: false,
		},
		{
			uri:     "",
			mustErr: true,
		},
		{
			uri:     "httpx//github.com",
			mustErr: true,
		},
		{
			uri:     "://github.com",
			mustErr: true,
		},
		{
			uri: "https://github.com/golang/go/blob/master/src/bytes/bytes_test.go",
			mustErr: false,
		},
	}

	for _, tc := range testCases {
		dst, err := New(tc.uri)
		if tc.mustErr {
			if err == nil {
				t.Errorf("%q should have err'd out", tc.uri)
			}
			if dst != nil {
				t.Errorf("expected a nil result, instead got %v for uri=%q", dst, tc.uri)
			}
		} else {
			if err != nil {
				t.Errorf("nil error expected %v; uri=%q", err, tc.uri)
			}
			if dst == nil {
				t.Errorf("expected non nil result for uri=%q", tc.uri)
			}
			if dst.Dirname() == "" {
				t.Errorf("expected a non-empty dirname() for uri=%q", tc.uri)
			}
			if dst.Basename() == "" {
				t.Errorf("expected a non-empty dirname() for uri=%q", tc.uri)
			}
			uriGot, uriWanted := dst.RawURI(), tc.uri
			if uriGot != uriWanted {
				t.Errorf("RawURI(): wanted %q, got %q", uriWanted, uriGot)
			}

			if dst.String() == "" {
				t.Errorf("expected a non-empty .String() result with uri=%q", tc.uri)
			}
			fmt.Printf("%q:: %q\n", dst.String(), tc.uri)
		}
	}
}

func BenchmarkURIToPathSt(b *testing.B) {
	candidates := []string{
		"file:///Users/emmanuelodeke/go/src/github.com/odeke-em/drive/src/copy.go",
		"https://github.com/odeke-em/drive/blob/master/src/copy.go",
		"",
		"https://github.com/golang/go/issues",
		"https://github.com/golang/go/blob/master/src/bytes/bytes_test.go",
		"https://",
	}

	for i, n := 0, b.N; i < n; i++ {
		for _, candidate := range candidates {
			dp, _ := New(candidate)
			if dp == nil {
			}
		}
	}
}
