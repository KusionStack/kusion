package storages

import (
	"testing"
)

func TestGenProjectDirPath(t *testing.T) {
	testCases := []struct {
		dir    string
		expect string
	}{
		{dir: "/home/user", expect: "/home/user/releases"},
	}

	for _, tc := range testCases {
		t.Run(tc.dir, func(t *testing.T) {
			actual := GenProjectDirPath(tc.dir)
			if actual != tc.expect {
				t.Errorf("GenProjectDirPath(%q) = %q; expect %q", tc.dir, actual, tc.expect)
			}
		})
	}
}

func TestGenGenericOssReleasePrefixKey(t *testing.T) {
	tests := []struct {
		prefix string
		want   string
	}{
		{"", releasesPrefix},
		{"/", releasesPrefix},
		{"some/prefix", "some/prefix/" + releasesPrefix},
		{"some_prefix", "some_prefix/" + releasesPrefix},
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			if got := GenGenericOssReleasePrefixKey(tt.prefix); got != tt.want {
				t.Errorf("GenGenericOssReleasePrefixKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
