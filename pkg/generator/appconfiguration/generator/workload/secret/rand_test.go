package secret

import (
	"strings"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	valid := "bcdfghjklmnpqrstvwxzBCDFGHJKLMNPQRSTVWXZ2456789"
	for _, l := range []int{0, 1, 2, 10, 52} {
		s := GenerateRandomString(l)
		if len(s) != l {
			t.Errorf("expected random string of size %d, actually got %q", l, s)
		}
		for _, c := range s {
			if !strings.ContainsRune(valid, c) {
				t.Errorf("expected valid characters, got %v", c)
			}
		}
	}
}

func BenchmarkRandomStringGeneration(b *testing.B) {
	b.ResetTimer()
	var s string
	for i := 0; i < b.N; i++ {
		s = GenerateRandomString(32)
	}
	b.StopTimer()
	if len(s) == 0 {
		b.Fatal(s)
	}
}
