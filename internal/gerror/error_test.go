package gerror

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestGetError(t *testing.T) {

	convey.Convey("GetError", t, func() {

		tests := []struct {
			desc     string
			input    APIErrorCode
			wantCode string
		}{
			{"success: ErrInvalidCopyDest", ErrInvalidCopyDest, "InvalidRequest"},
			{"error: not found", -99, "ErrNotFoundError"},
		}

		for _, test := range tests {

			convey.Convey(test.desc, func() {
				got := GetError(test.input, nil)

				convey.So(got.Code(), convey.ShouldEqual, test.wantCode)
			})
		}
	})
}

func TestGetAWSStdError(t *testing.T) {

	convey.Convey("GetAWSStdError", t, func() {
		tests := []struct {
			desc     string
			input    string
			wantCode string
		}{
			{"success: UserKeyMustBeSpecified", "UserKeyMustBeSpecified", "UserKeyMustBeSpecified"},
		}

		for _, test := range tests {

			convey.Convey(test.desc, func() {
				got := GetAWSStdError(test.input, nil)

				convey.So(got.Code(), convey.ShouldEqual, test.wantCode)
			})
		}
	})
}
