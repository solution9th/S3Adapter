package utils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCanonicalSQLKey(t *testing.T) {

	tests := []struct {
		input string
		want  string
	}{
		{"hello_world", "HelloWorld"},
		{"", ""},
		{"hello_", "Hello"},
		{"_hello", "Hello"},
		{"h_h_h_h", "HHHH"},
		{"._h", ".H"},
	}

	for k, v := range tests {
		if got := CanonicalSQLKey(v.input); got != v.want {
			t.Errorf("k: %v, in: %s, got: %v, want: %v\n", k, v.input, got, v.want)
		}
	}
}

func TestExists(t *testing.T) {

	convey.Convey("Exist", t, func() {

		tests := []struct {
			desc  string
			input string
			want  bool
		}{
			{"success", "utils_test.go", true},
			{"err: path not exist", "123.go", false},
			{"err: path is null", "", false},
		}

		for _, test := range tests {

			convey.Convey(test.desc, func() {
				got := Exists(test.input)

				convey.So(got, convey.ShouldEqual, test.want)
			})
		}
	})

}
