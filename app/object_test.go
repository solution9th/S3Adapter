package app

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/solution9th/S3Adapter/internal/gateway"
	"github.com/solution9th/S3Adapter/internal/gerror"
	"github.com/solution9th/S3Adapter/mocks/mock_gateway"

	"bou.ke/monkey"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/davecgh/go-spew/spew"
	"github.com/gavv/httpexpect"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
)

func TestHeadObject(t *testing.T) {

	convey.Convey("HeadObject", t, func() {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ex := newExpect(t, nil)

		tests := []struct {
			desc       string
			output     *s3.HeadObjectOutput
			resp       *http.Response
			gErr       error
			statusCode int
			hREQ       *httpexpect.Request
		}{
			{
				"success",
				&s3.HeadObjectOutput{},
				&http.Response{
					Header: map[string][]string{
						"X-Amz-Id-2":       []string{"P7uXNgaWsznGP2x/gUQEQ6gMxB69dByjkryXHWXZKEOaQcfueiVDc6E6cccP5AQ+2b+6XhtSGg4="},
						"X-Amz-Request-Id": []string{"E6D838D1B99BA637"},
						"Server":           []string{"AmazonS3"},
						"Date":             []string{"Wed, 05 Jun 2019 03:16:52 GMT"},
						"Content-Type":     []string{"application/xml"},
					},
				},
				nil,
				200,
				ex.HEAD("/bk/object.jpg"),
			},
			{
				"error: to UnmarshalRequest",
				&s3.HeadObjectOutput{},
				&http.Response{
					Header: map[string][]string{
						"X-Amz-Id-2":       []string{"P7uXNgaWsznGP2x/gUQEQ6gMxB69dByjkryXHWXZKEOaQcfueiVDc6E6cccP5AQ+2b+6XhtSGg4="},
						"X-Amz-Request-Id": []string{"E6D838D1B99BA637"},
						"Server":           []string{"AmazonS3"},
						"Date":             []string{"Wed, 05 Jun 2019 03:16:52 GMT"},
						"Content-Type":     []string{"application/xml"},
					},
				},
				nil,
				http.StatusOK,
				ex.HEAD("/bk/object.jpg").WithHeader("If-Unmodified-Since", "a"),
			},
			{
				"error: HeadObjectWithContext",
				&s3.HeadObjectOutput{},
				&http.Response{
					Header: map[string][]string{
						"X-Amz-Id-2":       []string{"P7uXNgaWsznGP2x/gUQEQ6gMxB69dByjkryXHWXZKEOaQcfueiVDc6E6cccP5AQ+2b+6XhtSGg4="},
						"X-Amz-Request-Id": []string{"E6D838D1B99BA637"},
						"Server":           []string{"AmazonS3"},
						"Date":             []string{"Wed, 05 Jun 2019 03:16:52 GMT"},
						"Content-Type":     []string{"application/xml"},
					},
				},
				gerror.GetError(gerror.ErrNoSuchKey, nil),
				http.StatusNotFound,
				ex.HEAD("/bk/object.jpg"),
			},
		}

		for _, test := range tests {

			convey.Convey(test.desc, func() {
				mockGateway := mock_gateway.NewMockS3Protocol(ctrl)

				mockGateway.EXPECT().HeadObjectWithContext(gomock.Any(), gomock.Any()).Return(test.output, test.resp, test.gErr).AnyTimes()

				// 控制变量
				var proto *API
				guard := monkey.PatchInstanceMethod(reflect.TypeOf(proto), "Auth", func(_ *API, _ *http.Request) bool {
					return true
				})
				defer guard.Unpatch()

				guard2 := monkey.PatchInstanceMethod(reflect.TypeOf(proto), "GetGateway", func(_ *API, _ *http.Request) gateway.S3Protocol {
					return mockGateway
				})
				defer guard2.Unpatch()

				e := test.hREQ.Expect().Status(test.statusCode)

				header := e.Headers().Raw()
				body := e.Body().Raw()

				spew.Dump(header)
				spew.Dump(body)

				// 重置
				ex = newExpect(t, nil)
			})
		}
	})
}

func TestGetObject(t *testing.T) {

	convey.Convey("GetObject", t, func() {

		return
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ex := newExpect(t, nil)

		tests := []struct {
			desc       string
			output     *s3.GetObjectOutput
			resp       *http.Response
			gErr       error
			statusCode int
			hREQ       *httpexpect.Request
		}{
			{
				"success",
				&s3.GetObjectOutput{},
				&http.Response{
					Header: map[string][]string{
						"X-Amz-Id-2":       []string{"P7uXNgaWsznGP2x/gUQEQ6gMxB69dByjkryXHWXZKEOaQcfueiVDc6E6cccP5AQ+2b+6XhtSGg4="},
						"X-Amz-Request-Id": []string{"E6D838D1B99BA637"},
						"Server":           []string{"AmazonS3"},
						"Date":             []string{"Wed, 05 Jun 2019 03:16:52 GMT"},
						"Content-Type":     []string{"application/xml"},
					},
				},
				nil,
				200,
				ex.HEAD("/bk/object.jpg"),
			},
			{
				"error: to UnmarshalRequest",
				&s3.GetObjectOutput{},
				&http.Response{
					Header: map[string][]string{
						"X-Amz-Id-2":       []string{"P7uXNgaWsznGP2x/gUQEQ6gMxB69dByjkryXHWXZKEOaQcfueiVDc6E6cccP5AQ+2b+6XhtSGg4="},
						"X-Amz-Request-Id": []string{"E6D838D1B99BA637"},
						"Server":           []string{"AmazonS3"},
						"Date":             []string{"Wed, 05 Jun 2019 03:16:52 GMT"},
						"Content-Type":     []string{"application/xml"},
					},
				},
				nil,
				http.StatusOK,
				ex.HEAD("/bk/object.jpg").WithHeader("If-Unmodified-Since", "a"),
			},
			{
				"error: HeadObjectWithContext",
				&s3.GetObjectOutput{},
				&http.Response{
					Header: map[string][]string{
						"X-Amz-Id-2":       []string{"P7uXNgaWsznGP2x/gUQEQ6gMxB69dByjkryXHWXZKEOaQcfueiVDc6E6cccP5AQ+2b+6XhtSGg4="},
						"X-Amz-Request-Id": []string{"E6D838D1B99BA637"},
						"Server":           []string{"AmazonS3"},
						"Date":             []string{"Wed, 05 Jun 2019 03:16:52 GMT"},
						"Content-Type":     []string{"application/xml"},
					},
				},
				gerror.GetError(gerror.ErrNoSuchKey, nil),
				http.StatusNotFound,
				ex.HEAD("/bk/object.jpg"),
			},
		}

		for _, test := range tests {

			convey.Convey(test.desc, func() {
				mockGateway := mock_gateway.NewMockS3Protocol(ctrl)

				mockGateway.EXPECT().GetObjectWithContext(gomock.Any(), gomock.Any()).Return(test.output, test.resp, test.gErr).AnyTimes()

				// 控制变量
				var proto *API
				guard := monkey.PatchInstanceMethod(reflect.TypeOf(proto), "Auth", func(_ *API, _ *http.Request) bool {
					return true
				})
				defer guard.Unpatch()

				guard2 := monkey.PatchInstanceMethod(reflect.TypeOf(proto), "GetGateway", func(_ *API, _ *http.Request) gateway.S3Protocol {
					return mockGateway
				})
				defer guard2.Unpatch()

				e := test.hREQ.Expect().Status(test.statusCode)

				header := e.Headers().Raw()
				body := e.Body().Raw()

				spew.Dump(header)
				spew.Dump(body)

				// 重置
				ex = newExpect(t, nil)
			})
		}
	})
}

func TestDeleteObject(t *testing.T) {

	convey.Convey("DeleteObject", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ex := newExpect(t, nil)

		tests := []struct {
			desc       string
			output     *s3.DeleteObjectOutput
			resp       *http.Response
			gErr       error
			statusCode int
			hREQ       *httpexpect.Request
		}{
			{
				"success",
				&s3.DeleteObjectOutput{},
				&http.Response{
					Header: map[string][]string{
						"X-Amz-Id-2":       []string{"P7uXNgaWsznGP2x/gUQEQ6gMxB69dByjkryXHWXZKEOaQcfueiVDc6E6cccP5AQ+2b+6XhtSGg4="},
						"X-Amz-Request-Id": []string{"E6D838D1B99BA637"},
						"Server":           []string{"AmazonS3"},
						"Date":             []string{"Wed, 05 Jun 2019 03:16:52 GMT"},
						"Content-Type":     []string{"application/xml"},
					},
				},
				nil,
				http.StatusOK,
				ex.DELETE("/bk/object.jpg"),
			},
			{
				"error: to UnmarshalRequest",
				&s3.DeleteObjectOutput{},
				&http.Response{
					Header: map[string][]string{
						"X-Amz-Id-2":       []string{"P7uXNgaWsznGP2x/gUQEQ6gMxB69dByjkryXHWXZKEOaQcfueiVDc6E6cccP5AQ+2b+6XhtSGg4="},
						"X-Amz-Request-Id": []string{"E6D838D1B99BA637"},
						"Server":           []string{"AmazonS3"},
						"Date":             []string{"Wed, 05 Jun 2019 03:16:52 GMT"},
						"Content-Type":     []string{"application/xml"},
						"MFA":              []string{""},
					},
				},
				nil,
				http.StatusOK,
				ex.DELETE("/bk/object.jpg"),
			},
			{
				"error: DeleteObjectWithContext",
				&s3.DeleteObjectOutput{},
				&http.Response{
					Header: map[string][]string{
						"X-Amz-Id-2":       []string{"P7uXNgaWsznGP2x/gUQEQ6gMxB69dByjkryXHWXZKEOaQcfueiVDc6E6cccP5AQ+2b+6XhtSGg4="},
						"X-Amz-Request-Id": []string{"E6D838D1B99BA637"},
						"Server":           []string{"AmazonS3"},
						"Date":             []string{"Wed, 05 Jun 2019 03:16:52 GMT"},
						"Content-Type":     []string{"application/xml"},
					},
				},
				gerror.GetError(gerror.ErrNoSuchKey, nil),
				http.StatusNotFound,
				ex.DELETE("/bk/object.jpg"),
			},
		}

		for _, test := range tests {

			convey.Convey(test.desc, func() {

				mockGateway := mock_gateway.NewMockS3Protocol(ctrl)

				mockGateway.EXPECT().DeleteObjectWithContext(gomock.Any(), gomock.Any()).Return(test.output, test.resp, test.gErr).AnyTimes()

				// 控制变量
				var proto *API
				guard := monkey.PatchInstanceMethod(reflect.TypeOf(proto), "Auth", func(_ *API, _ *http.Request) bool {
					return true
				})
				defer guard.Unpatch()

				guard2 := monkey.PatchInstanceMethod(reflect.TypeOf(proto), "GetGateway", func(_ *API, _ *http.Request) gateway.S3Protocol {
					return mockGateway
				})
				defer guard2.Unpatch()

				e := test.hREQ.Expect().Status(test.statusCode)

				header := e.Headers().Raw()
				body := e.Body().Raw()

				spew.Dump(header)
				spew.Dump(body)

				// 重置
				ex = newExpect(t, nil)
			})
		}
	})
}

func TestCopyObject(t *testing.T) {

	convey.Convey("CopyObject", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ex := newExpect(t, nil)

		tests := []struct {
			desc       string
			output     *s3.CopyObjectOutput
			resp       *http.Response
			gErr       error
			statusCode int
			hREQ       *httpexpect.Request
		}{
			{
				"success",
				&s3.CopyObjectOutput{},
				&http.Response{
					Header: map[string][]string{
						"X-Amz-Id-2":       []string{"P7uXNgaWsznGP2x/gUQEQ6gMxB69dByjkryXHWXZKEOaQcfueiVDc6E6cccP5AQ+2b+6XhtSGg4="},
						"X-Amz-Request-Id": []string{"E6D838D1B99BA637"},
						"Server":           []string{"AmazonS3"},
						"Date":             []string{"Wed, 05 Jun 2019 03:16:52 GMT"},
						"Content-Type":     []string{"application/xml"},
					},
				},
				nil,
				http.StatusOK,
				ex.PUT("/bk/object.jpg").WithHeader("X-Amz-Copy-Source", "bk/source.jpg"),
			},

			{
				"error: CopyObjectWithContext",
				&s3.CopyObjectOutput{},
				&http.Response{
					Header: map[string][]string{
						"X-Amz-Id-2":        []string{"P7uXNgaWsznGP2x/gUQEQ6gMxB69dByjkryXHWXZKEOaQcfueiVDc6E6cccP5AQ+2b+6XhtSGg4="},
						"X-Amz-Request-Id":  []string{"E6D838D1B99BA637"},
						"Server":            []string{"AmazonS3"},
						"Date":              []string{"Wed, 05 Jun 2019 03:16:52 GMT"},
						"Content-Type":      []string{"application/xml"},
						"X-Amz-Copy-Source": []string{"123"},
					},
				},
				gerror.GetError(gerror.ErrNoSuchKey, nil),
				http.StatusNotFound,
				ex.PUT("/bk/object.jpg").WithHeader("X-Amz-Copy-Source", "bk/source.jpg"),
			},
		}

		for _, test := range tests {

			convey.Convey(test.desc, func() {

				mockGateway := mock_gateway.NewMockS3Protocol(ctrl)

				mockGateway.EXPECT().CopyObjectWithContext(gomock.Any(), gomock.Any()).Return(test.output, test.resp, test.gErr).AnyTimes()

				// 控制变量
				var proto *API
				guard := monkey.PatchInstanceMethod(reflect.TypeOf(proto), "Auth", func(_ *API, _ *http.Request) bool {
					return true
				})
				defer guard.Unpatch()

				guard2 := monkey.PatchInstanceMethod(reflect.TypeOf(proto), "GetGateway", func(_ *API, _ *http.Request) gateway.S3Protocol {
					return mockGateway
				})
				defer guard2.Unpatch()

				e := test.hREQ.Expect().Status(test.statusCode)

				header := e.Headers().Raw()
				body := e.Body().Raw()

				spew.Dump(header)
				spew.Dump(body)

				// 重置
				ex = newExpect(t, nil)
			})
		}
	})
}
