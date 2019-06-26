package sign

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/solution9th/S3Adapter/internal/gerror"

	"github.com/smartystreets/goconvey/convey"
)

func TestGetURLParams(t *testing.T) {

	convey.Convey("TestGetURLParams", t, func() {

		tests := []struct {
			desc string
			ak   string
			sk   string
			r    *http.Request
			want gerror.APIErrorCode
		}{
			{
				"success",
				"AKIAIOSFODNN7EXAMPLE",
				"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				httptest.NewRequest("GET", "https://examplebucket.s3.amazonaws.com/test.txt?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAIOSFODNN7EXAMPLE%2F20130524%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20130524T000000Z&X-Amz-Expires=86400&X-Amz-SignedHeaders=host&X-Amz-Signature=aeeed9bbccd4d02ee5c0109b86d86835f995330da4c265957d157751f604d404", nil),
				gerror.ErrNone,
			},
		}

		for _, test := range tests {

			convey.Convey(test.desc, func() {

				Debug = true
				s := NewSignV4(test.ak, test.sk, "us-east-1", test.r)

				utc, _ := time.Parse(http.TimeFormat, "Fri, 24 May 2013 00:00:00 GMT")
				errCode := s.VerifyURL(utc)

				convey.So(errCode, convey.ShouldEqual, test.want)

			})
		}

	})
}
