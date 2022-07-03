package config

import (
	"os"
	"testing"
)

func TestExpandValueEnv(t *testing.T) {
	testCases := []struct {
		item string
		want string
	}{
		{"", ""},
		{"$", "$"},
		{"{", "{"},
		{"{}", "{}"},
		{"${}", ""},
		{"${|}", ""},
		{"${}", ""},
		{"${{}}", ""},
		{"${{||}}", "}"},
		{"${pwd||}", ""},
		{"${pwd||}", ""},
		{"${pwd||}", ""},
		{"${pwd||}}", "}"},
		{"${pwd||{{||}}}", "{{||}}"},
		{"${GOPATH}", os.Getenv("GOPATH")},
		{"${GOPATH||}", os.Getenv("GOPATH")},
		{"${GOPATH||root}", os.Getenv("GOPATH")},
		{"${GOPATH_NOT||root}", "root"},
		{"${GOPATH_NOT||||root}", "||root"},
	}

	for _, c := range testCases {
		if got := ExpandValueEnv(c.item); got != c.want {
			t.Errorf("expand value error, item %q want %q, got %q", c.item, c.want, got)
		}
	}
}
