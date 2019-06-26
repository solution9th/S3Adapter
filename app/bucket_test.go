package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/solution9th/S3Adapter/internal/db"
	"github.com/solution9th/S3Adapter/internal/db/mysql"
	"github.com/solution9th/S3Adapter/internal/gateway"
	"github.com/solution9th/S3Adapter/internal/gerror"
	"github.com/solution9th/S3Adapter/mocks/mock_db"
	"github.com/solution9th/S3Adapter/mocks/mock_gateway"

	"bou.ke/monkey"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/davecgh/go-spew/spew"
	"github.com/gavv/httpexpect"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/smartystreets/goconvey/convey"
)

// func init() {
// 	zlog.NoColor = true
// 	zlog.NewBasicLog(os.Stdout)
// }

func newExpect(t *testing.T, db db.DB) *httpexpect.Expect {
	r := mux.NewRouter()
	NewAPIRouter(r, &API{
		DB: db,
	})
	server := httptest.NewServer(r)
	return httpexpect.New(t, server.URL)
}

func TestPutApplication(t *testing.T) {

	convey.Convey("PutApplication", t, func() {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		convey.Convey("success: all success", func() {

			mockDB := mock_db.NewMockDB(ctrl)
			mockDB.EXPECT().CountInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(0, nil)
			mockDB.EXPECT().SaveInfo(gomock.Any()).Return(1, nil)

			e := newExpect(t, mockDB)

			e.PUT("/").WithBytes([]byte("<CreateApplicationConfiguration><AccessKey>123</AccessKey><SecretKey>123</SecretKey><Engine>s3</Engine><AppName>Test</AppName><AppRemark>Test</AppRemark></CreateApplicationConfiguration>")).Expect().Status(200)
		})

		convey.Convey("err: xml decode", func() {
			e := newExpect(t, nil)

			e.PUT("/").WithBytes([]byte("<>")).Expect().Status(http.StatusBadRequest)
		})

		convey.Convey("err: saveInfo db count info", func() {
			mockDB := mock_db.NewMockDB(ctrl)
			mockDB.EXPECT().CountInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(0, fmt.Errorf("db: count info error"))

			e := newExpect(t, mockDB)

			e.PUT("/").WithBytes([]byte("<CreateApplicationConfiguration><AccessKey>123</AccessKey><SecretKey>123</SecretKey><Engine>s3</Engine><AppName>Test</AppName><AppRemark>Test</AppRemark></CreateApplicationConfiguration>")).Expect().Status(http.StatusServiceUnavailable)
		})

		convey.Convey("err: saveInfo num > 1", func() {
			mockDB := mock_db.NewMockDB(ctrl)
			mockDB.EXPECT().CountInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(1, nil)

			e := newExpect(t, mockDB)

			e.PUT("/").WithBytes([]byte("<CreateApplicationConfiguration><AccessKey>123</AccessKey><SecretKey>123</SecretKey><Engine>s3</Engine><AppName>Test</AppName><AppRemark>Test</AppRemark></CreateApplicationConfiguration>")).Expect().Status(http.StatusServiceUnavailable)
		})

		convey.Convey("err: saveInfo db saveInfo", func() {
			mockDB := mock_db.NewMockDB(ctrl)
			mockDB.EXPECT().CountInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(0, nil)
			mockDB.EXPECT().SaveInfo(gomock.Any()).Return(0, fmt.Errorf("db: save info error"))

			e := newExpect(t, mockDB)

			e.PUT("/").WithBytes([]byte("<CreateApplicationConfiguration><AccessKey>123</AccessKey><SecretKey>123</SecretKey><Engine>s3</Engine><AppName>Test</AppName><AppRemark>Test</AppRemark></CreateApplicationConfiguration>")).Expect().Status(http.StatusServiceUnavailable)
		})

	})
}

func TestDeleteApplication(t *testing.T) {

	convey.Convey("DeleteApplication", t, func() {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		convey.Convey("success: all success", func() {

			mockDB := mock_db.NewMockDB(ctrl)
			mockDB.EXPECT().GetInfo(gomock.Any()).Return(mysql.Info{
				OsAccessKey:     "oak",
				OsScrectKey:     "osk",
				EngineAccessKey: "akak",
				EngineSecretKey: "sksk",
				EngineType:      "s3",
			}, nil).AnyTimes()
			mockDB.EXPECT().DeleteInfo(gomock.Any(), gomock.Any()).Return(nil)

			e := newExpect(t, mockDB)

			// 暂时签名验证还没有时间限制
			e.DELETE("/").WithHeader("Authorization", "AWS4-HMAC-SHA256 Credential=akak/20190604/beijing/s3/aws4_request, SignedHeaders=date;host, Signature=eee631d97e7b1de5b8d4e4bb5354eeba1d0be2c1b83d3f200da54ff65fff2cb5").WithHeader("Date", "Tue, 04 Jun 2019 14:51:19 GMT").WithHeader("Host", "s3.newio.cc").Expect().Status(http.StatusOK)
		})

		convey.Convey("err: auth is nil", func() {

			mockDB := mock_db.NewMockDB(ctrl)
			mockDB.EXPECT().GetInfo(gomock.Any()).Return(mysql.Info{
				OsAccessKey:     "oak",
				OsScrectKey:     "osk",
				EngineAccessKey: "akak",
				EngineSecretKey: "sksk",
				EngineType:      "s3",
			}, nil).AnyTimes()
			// mockDB.EXPECT().DeleteInfo(gomock.Any(), gomock.Any()).Return(nil)

			e := newExpect(t, mockDB)

			e.DELETE("/").Expect().Status(http.StatusForbidden)
		})

		convey.Convey("err: authsign not has AWS4HMACSHA256", func() {

			mockDB := mock_db.NewMockDB(ctrl)
			mockDB.EXPECT().GetInfo(gomock.Any()).Return(mysql.Info{
				OsAccessKey:     "oak",
				OsScrectKey:     "osk",
				EngineAccessKey: "akak",
				EngineSecretKey: "sksk",
				EngineType:      "s3",
			}, nil).AnyTimes()

			e := newExpect(t, mockDB)
			e.DELETE("/").WithHeader("Authorization", "123").Expect().Status(http.StatusForbidden)
		})

		convey.Convey("err: authsign format error", func() {

			mockDB := mock_db.NewMockDB(ctrl)
			mockDB.EXPECT().GetInfo(gomock.Any()).Return(mysql.Info{
				OsAccessKey:     "oak",
				OsScrectKey:     "osk",
				EngineAccessKey: "akak",
				EngineSecretKey: "sksk",
				EngineType:      "s3",
			}, nil).AnyTimes()

			e := newExpect(t, mockDB)

			e.DELETE("/").WithHeader("Authorization", "AWS4-HMAC-SHA256 456").Expect().Status(http.StatusForbidden)
		})

		convey.Convey("err: authsign format error content", func() {

			mockDB := mock_db.NewMockDB(ctrl)
			mockDB.EXPECT().GetInfo(gomock.Any()).Return(mysql.Info{
				OsAccessKey:     "oak",
				OsScrectKey:     "osk",
				EngineAccessKey: "akak",
				EngineSecretKey: "sksk",
				EngineType:      "s3",
			}, nil).AnyTimes()

			e := newExpect(t, mockDB)

			// Credential is nil
			e.DELETE("/").WithHeader("Authorization", "AWS4-HMAC-SHA256 Credential=, SignedHeaders=host;range;x-amz-date, Signature=fe5f80f77d5fa3beca038a248ff027d0445342fe2855ddc963176630326f1024").Expect().Status(http.StatusForbidden)
		})

		convey.Convey("err: sign verify success but delete error", func() {

			mockDB := mock_db.NewMockDB(ctrl)
			mockDB.EXPECT().GetInfo(gomock.Any()).Return(mysql.Info{
				OsAccessKey:     "oak",
				OsScrectKey:     "osk",
				EngineAccessKey: "akak",
				EngineSecretKey: "sksk",
				EngineType:      "s3",
			}, nil).AnyTimes()
			mockDB.EXPECT().DeleteInfo(gomock.Any(), gomock.Any()).Return(fmt.Errorf("db: delete error"))

			e := newExpect(t, mockDB)

			// 暂时签名验证还没有时间限制
			e.DELETE("/").WithHeader("Authorization", "AWS4-HMAC-SHA256 Credential=akak/20190604/beijing/s3/aws4_request, SignedHeaders=date;host, Signature=eee631d97e7b1de5b8d4e4bb5354eeba1d0be2c1b83d3f200da54ff65fff2cb5").WithHeader("Date", "Tue, 04 Jun 2019 14:51:19 GMT").WithHeader("Host", "s3.newio.cc").Expect().Status(http.StatusServiceUnavailable)
		})

		convey.Convey("err: db get info error", func() {

			mockDB := mock_db.NewMockDB(ctrl)
			mockDB.EXPECT().GetInfo(gomock.Any()).Return(mysql.Info{}, fmt.Errorf("db: get info")).AnyTimes()

			e := newExpect(t, mockDB)

			// 暂时签名验证还没有时间限制
			e.DELETE("/").WithHeader("Authorization", "AWS4-HMAC-SHA256 Credential=akak/20190604/beijing/s3/aws4_request, SignedHeaders=date;host, Signature=eee631d97e7b1de5b8d4e4bb5354eeba1d0be2c1b83d3f200da54ff65fff2cb5").WithHeader("Date", "Tue, 04 Jun 2019 14:51:19 GMT").WithHeader("Host", "s3.newio.cc").Expect().Status(http.StatusForbidden)
		})
	})
}

func TestHeadBucket(t *testing.T) {

	convey.Convey("HeadBucket", t, func() {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		convey.Convey("success: all success", func() {
			mockDB := mock_db.NewMockDB(ctrl)
			mockDB.EXPECT().GetInfo(gomock.Any()).Return(mysql.Info{
				OsAccessKey:     "oak",
				OsScrectKey:     "osk",
				EngineAccessKey: "akak",
				EngineSecretKey: "sksk",
				EngineType:      "s3",
			}, nil).AnyTimes()

			resp := &http.Response{
				StatusCode: http.StatusOK,
			}

			mockGateway := mock_gateway.NewMockS3Protocol(ctrl)
			mockGateway.EXPECT().HeadBucketWithContext(gomock.Any(), gomock.Any()).Return(nil, resp, nil).AnyTimes()

			var proto *API

			monkey.PatchInstanceMethod(reflect.TypeOf(proto), "GetGateway", func(_ *API, _ *http.Request) gateway.S3Protocol {
				fmt.Println("=> return mockGateway")
				return mockGateway
			})

			e := newExpect(t, mockDB)

			e.HEAD("/bkbk").WithHeader("Authorization", "AWS4-HMAC-SHA256 Credential=akak/20190604/beijing/s3/aws4_request, SignedHeaders=date;host, Signature=f1c1c51a146d0648f6607f3a23bc82cbdc2fe17039d19b4bd1a2a86c9a90d779").WithHeader("Date", "Tue, 04 Jun 2019 14:51:19 GMT").WithHeader("Host", "s3.newio.cc").Expect().Status(http.StatusOK)
		})

		convey.Convey("err: gateway not found", func() {
			mockDB := mock_db.NewMockDB(ctrl)
			mockDB.EXPECT().GetInfo(gomock.Any()).Return(mysql.Info{
				OsAccessKey:     "oak",
				OsScrectKey:     "osk",
				EngineAccessKey: "akak",
				EngineSecretKey: "sksk",
				EngineType:      "s3",
			}, nil).AnyTimes()

			var proto *API

			monkey.PatchInstanceMethod(reflect.TypeOf(proto), "GetGateway", func(_ *API, _ *http.Request) gateway.S3Protocol {
				fmt.Println("=> return gateway not found")
				return nil
			})

			e := newExpect(t, mockDB)

			ee := e.HEAD("/bkbk").WithHeader("Authorization", "AWS4-HMAC-SHA256 Credential=akak/20190604/beijing/s3/aws4_request, SignedHeaders=date;host, Signature=f1c1c51a146d0648f6607f3a23bc82cbdc2fe17039d19b4bd1a2a86c9a90d779").WithHeader("Date", "Tue, 04 Jun 2019 14:51:19 GMT").WithHeader("Host", "s3.newio.cc").Expect()

			ee.Status(http.StatusServiceUnavailable)
		})

		convey.Convey("err: HeadBucketWithContext error", func() {
			mockDB := mock_db.NewMockDB(ctrl)
			mockDB.EXPECT().GetInfo(gomock.Any()).Return(mysql.Info{
				OsAccessKey:     "AKIAIYXNSBM66BIOC3TQ",
				OsScrectKey:     "CxNHhX4qUmbDizUZpmxRjUDzyllNPvZicmwG54t6",
				EngineAccessKey: "AKIAIYXNSBM66BIOC3TQ",
				EngineSecretKey: "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",
				EngineType:      "s3",
			}, nil).AnyTimes()

			headerMap := map[string]string{
				"X-Amz-Id-2":       "P7uXNgaWsznGP2x/gUQEQ6gMxB69dByjkryXHWXZKEOaQcfueiVDc6E6cccP5AQ+2b+6XhtSGg4=",
				"X-Amz-Request-Id": "E6D838D1B99BA637",
				"Server":           "AmazonS3",
				"Date":             "Wed, 05 Jun 2019 03:16:52 GMT",
			}

			tHeader := make(http.Header)

			for k, v := range headerMap {
				tHeader.Add(k, v)
			}

			resp := &http.Response{
				StatusCode: http.StatusNotFound,
				Header:     tHeader,
			}

			mockGateway := mock_gateway.NewMockS3Protocol(ctrl)
			mockGateway.EXPECT().HeadBucketWithContext(gomock.Any(), gomock.Any()).Return(nil, resp, gerror.GetError(gerror.ErrNoSuchBucket, nil)).AnyTimes()

			var proto *API

			guard := monkey.PatchInstanceMethod(reflect.TypeOf(proto), "GetGateway", func(_ *API, _ *http.Request) gateway.S3Protocol {
				fmt.Println("=> return mockGateway")
				return mockGateway
			})
			defer guard.Unpatch()

			e := newExpect(t, mockDB)

			e.HEAD("/TestS3TestHeadBucket").WithHeader("Authorization", "AWS4-HMAC-SHA256 Credential=AKIAIYXNSBM66BIOC3TQ/20190605/us-east-1/s3/aws4_request, SignedHeaders=host;x-amz-content-sha256;x-amz-date, Signature=abf7104560707a60923dd38d1bbeef8e5043820eda2012accbce017baf1955bd").WithHeader("Host", "s3.amazonaws.com").WithHeader("X-Amz-Content-Sha256", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855").WithHeader("X-Amz-Date", "20190605T031650Z").Expect().Status(http.StatusNotFound)

		})

	})
}

// Auth 和 Gateway

func TestGetBucketV1(t *testing.T) {

	convey.Convey("GetBucketV1", t, func() {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ex := newExpect(t, nil)

		tests := []struct {
			desc       string
			output     *s3.ListObjectsOutput
			resp       *http.Response
			gErr       error
			statusCode int
			hREQ       *httpexpect.Request
		}{
			{
				"success",
				&s3.ListObjectsOutput{},
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
				ex.GET("/bkname").WithQuery("encoding-type", "url"),
			},
			{
				"error: max-keys not number",
				&s3.ListObjectsOutput{},
				&http.Response{},
				nil,
				http.StatusBadRequest,
				ex.GET("/bkname").WithQuery("max-keys", "a"),
			},
			{
				"error: encoding-type isn't url",
				&s3.ListObjectsOutput{},
				&http.Response{},
				nil,
				http.StatusBadRequest,
				ex.GET("/bkname").WithQuery("encoding-type", "a"),
			},
			{
				"error: ListObjectsWithContext",
				&s3.ListObjectsOutput{},
				&http.Response{
					Header: map[string][]string{
						"X-Amz-Id-2":       []string{"P7uXNgaWsznGP2x/gUQEQ6gMxB69dByjkryXHWXZKEOaQcfueiVDc6E6cccP5AQ+2b+6XhtSGg4="},
						"X-Amz-Request-Id": []string{"E6D838D1B99BA637"},
						"Server":           []string{"AmazonS3"},
						"Date":             []string{"Wed, 05 Jun 2019 03:16:52 GMT"},
						"Content-Type":     []string{"application/xml"},
					},
				},
				gerror.GetError(gerror.ErrInvalidEncodingMethod, nil),
				http.StatusBadRequest,
				ex.GET("/bkname"),
			},
		}

		for _, test := range tests {

			convey.Convey(test.desc, func() {
				mockGateway := mock_gateway.NewMockS3Protocol(ctrl)

				mockGateway.EXPECT().ListObjectsWithContext(gomock.Any(), gomock.Any()).Return(test.output, test.resp, test.gErr).AnyTimes()

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

func TestGetBucketV2(t *testing.T) {

	convey.Convey("GetBucketV2", t, func() {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ex := newExpect(t, nil)

		tests := []struct {
			desc       string
			output     *s3.ListObjectsV2Output
			resp       *http.Response
			gErr       error
			statusCode int
			hREQ       *httpexpect.Request
		}{
			{
				"success",
				&s3.ListObjectsV2Output{},
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
				ex.GET("/bkname").WithQuery("list-type", "2").WithQuery("encoding-type", "url"),
			},
			{
				"error: continuation-token is nil",
				&s3.ListObjectsV2Output{},
				&http.Response{},
				nil,
				http.StatusBadRequest,
				ex.GET("/bkname").WithQuery("list-type", "2").WithQuery("continuation-token", ""),
			},
			{
				"error: max-keys not number",
				&s3.ListObjectsV2Output{},
				&http.Response{},
				nil,
				http.StatusBadRequest,
				ex.GET("/bkname").WithQuery("list-type", "2").WithQuery("max-keys", "a"),
			},
			{
				"error: encoding-type isn't url",
				&s3.ListObjectsV2Output{},
				&http.Response{},
				nil,
				http.StatusBadRequest,
				ex.GET("/bkname").WithQuery("list-type", "2").WithQuery("encoding-type", "a"),
			},
			{
				"error: ListObjectsWithContextV2",
				&s3.ListObjectsV2Output{},
				&http.Response{
					Header: map[string][]string{
						"X-Amz-Id-2":       []string{"P7uXNgaWsznGP2x/gUQEQ6gMxB69dByjkryXHWXZKEOaQcfueiVDc6E6cccP5AQ+2b+6XhtSGg4="},
						"X-Amz-Request-Id": []string{"E6D838D1B99BA637"},
						"Server":           []string{"AmazonS3"},
						"Date":             []string{"Wed, 05 Jun 2019 03:16:52 GMT"},
						"Content-Type":     []string{"application/xml"},
					},
				},
				gerror.GetError(gerror.ErrInvalidEncodingMethod, nil),
				http.StatusBadRequest,
				ex.GET("/bkname").WithQuery("list-type", "2"),
			},
		}

		for _, test := range tests {

			convey.Convey(test.desc, func() {
				mockGateway := mock_gateway.NewMockS3Protocol(ctrl)

				mockGateway.EXPECT().ListObjectsWithContextV2(gomock.Any(), gomock.Any()).Return(test.output, test.resp, test.gErr).AnyTimes()

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

func TestListBuckets(t *testing.T) {

	convey.Convey("ListBuckets", t, func() {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ex := newExpect(t, nil)

		tests := []struct {
			desc       string
			output     *s3.ListBucketsOutput
			resp       *http.Response
			gErr       error
			statusCode int
			hREQ       *httpexpect.Request
		}{
			{
				"success",
				&s3.ListBucketsOutput{},
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
				ex.GET("/"),
			},
			{
				"success: with output",
				&s3.ListBucketsOutput{
					Buckets: []*s3.Bucket{
						{
							Name:         aws.String("bk"),
							CreationDate: aws.Time(time.Now()),
						},
					},
					Owner: &s3.Owner{
						DisplayName: aws.String("name"),
						ID:          aws.String("id"),
					},
				},
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
				ex.GET("/"),
			},
			{
				"error: ListBucketsWithContext",
				&s3.ListBucketsOutput{},
				&http.Response{
					Header: map[string][]string{
						"X-Amz-Id-2":       []string{"P7uXNgaWsznGP2x/gUQEQ6gMxB69dByjkryXHWXZKEOaQcfueiVDc6E6cccP5AQ+2b+6XhtSGg4="},
						"X-Amz-Request-Id": []string{"E6D838D1B99BA637"},
						"Server":           []string{"AmazonS3"},
						"Date":             []string{"Wed, 05 Jun 2019 03:16:52 GMT"},
						"Content-Type":     []string{"application/xml"},
					},
				},
				gerror.GetError(gerror.ErrInvalidEncodingMethod, nil),
				http.StatusBadRequest,
				ex.GET("/"),
			},
		}

		for _, test := range tests {

			convey.Convey(test.desc, func() {
				mockGateway := mock_gateway.NewMockS3Protocol(ctrl)

				mockGateway.EXPECT().ListBucketsWithContext(gomock.Any(), gomock.Any()).Return(test.output, test.resp, test.gErr).AnyTimes()

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

func TestDeleteBucket(t *testing.T) {

	convey.Convey("DeleteBucket", t, func() {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ex := newExpect(t, nil)

		tests := []struct {
			desc       string
			output     *s3.DeleteBucketOutput
			resp       *http.Response
			gErr       error
			statusCode int
			hREQ       *httpexpect.Request
		}{
			{
				"success",
				&s3.DeleteBucketOutput{},
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
				ex.DELETE("/bkbk"),
			},
			{
				"error: DeleteBucketWithContext",
				&s3.DeleteBucketOutput{},
				&http.Response{
					Header: map[string][]string{
						"X-Amz-Id-2":       []string{"P7uXNgaWsznGP2x/gUQEQ6gMxB69dByjkryXHWXZKEOaQcfueiVDc6E6cccP5AQ+2b+6XhtSGg4="},
						"X-Amz-Request-Id": []string{"E6D838D1B99BA637"},
						"Server":           []string{"AmazonS3"},
						"Date":             []string{"Wed, 05 Jun 2019 03:16:52 GMT"},
						"Content-Type":     []string{"application/xml"},
					},
				},
				gerror.GetError(gerror.ErrInvalidEncodingMethod, nil),
				http.StatusBadRequest,
				ex.DELETE("/bkbk"),
			},
		}

		for _, test := range tests {

			convey.Convey(test.desc, func() {
				mockGateway := mock_gateway.NewMockS3Protocol(ctrl)

				mockGateway.EXPECT().DeleteBucketWithContext(gomock.Any(), gomock.Any()).Return(test.output, test.resp, test.gErr).AnyTimes()

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

func TestPutBucket(t *testing.T) {

	convey.Convey("PutBucket", t, func() {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ex := newExpect(t, nil)

		tests := []struct {
			desc       string
			output     *s3.CreateBucketOutput
			resp       *http.Response
			gErr       error
			statusCode int
			hREQ       *httpexpect.Request
		}{
			{
				"success",
				&s3.CreateBucketOutput{},
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
				ex.PUT("/bkbk"),
			},
			{
				"success: location is not nil",
				&s3.CreateBucketOutput{
					Location: aws.String("BucketRegion"),
				},
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
				ex.PUT("/bkbk").WithBytes([]byte(`<CreateBucketConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/"> <LocationConstraint>BucketRegion</LocationConstraint></CreateBucketConfiguration>`)),
			},
			{
				"error: xml error",
				&s3.CreateBucketOutput{},
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
				http.StatusBadRequest,
				ex.PUT("/bkbk").WithBytes([]byte(`<>`)),
			},
			{
				"error: DeleteBucketWithContext",
				&s3.CreateBucketOutput{},
				&http.Response{
					Header: map[string][]string{
						"X-Amz-Id-2":       []string{"P7uXNgaWsznGP2x/gUQEQ6gMxB69dByjkryXHWXZKEOaQcfueiVDc6E6cccP5AQ+2b+6XhtSGg4="},
						"X-Amz-Request-Id": []string{"E6D838D1B99BA637"},
						"Server":           []string{"AmazonS3"},
						"Date":             []string{"Wed, 05 Jun 2019 03:16:52 GMT"},
						"Content-Type":     []string{"application/xml"},
					},
				},
				gerror.GetError(gerror.ErrInvalidEncodingMethod, nil),
				http.StatusBadRequest,
				ex.PUT("/bkbk"),
			},
		}

		for _, test := range tests {

			convey.Convey(test.desc, func() {
				mockGateway := mock_gateway.NewMockS3Protocol(ctrl)

				mockGateway.EXPECT().CreateBucketWithContext(gomock.Any(), gomock.Any()).Return(test.output, test.resp, test.gErr).AnyTimes()

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
