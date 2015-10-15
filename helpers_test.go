package template

import "testing"

func TestStripExt(t *testing.T) {
	tests := [][]string{
		{"abc.ext", "abc"},
		{"abc.tar.gz", "abc"},
		{"a.b.c/abc.tar.gz", "a.b.c/abc"},
		{"a.b.c/.", "a.b.c/"},
		{"a.b.c/", "a.b.c/"},
	}

	for _, test := range tests {
		if r := stripExt(test[0]); r != test[1] {
			t.Errorf("got\n\t%+v\nexpected\n\t%v", r, test[1])
		}
	}
}
