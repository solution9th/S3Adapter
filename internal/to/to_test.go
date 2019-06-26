package to

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/davecgh/go-spew/spew"
)

func TestUnmarshalRequest(t *testing.T) {

	tests := []struct {
		r       *http.Request
		header  map[string]string
		payload string
		data    interface{}
		want    error
	}{
		{
			httptest.NewRequest("GET", "http://newio.cc", nil),
			map[string]string{
				"If-Match":            "vIf-Match",
				"If-Modified-Since":   time.Now().Format(http.TimeFormat),
				"If-None-Match":       "vIf-None-Match",
				"If-Unmodified-Since": time.Now().Format(http.TimeFormat),
				"Range":               "vRange",
				"x-amz-request-payer": "vx-amz-request-payer",
				"x-amz-server-side-encryption-customer-algorithm": "vx-amz-server-side-encryption-customer-algorithm",
			},
			"",
			&s3.GetObjectInput{},
			nil,
		},
		{
			httptest.NewRequest("GET", "http://newio.cc", nil),
			map[string]string{
				"x-amz-grant-write-acp": "vx-amz-grant-write-acp",
				"x-amz-meta-zero":       "x-amz-meta-0",
				"x-amz-meta-one":        "x-amz-meta-1",
				"x-amz-meta-two":        "x-amz-meta-2",
			},
			"",
			&s3.CreateMultipartUploadInput{},
			nil,
		},
		{
			httptest.NewRequest("GET", "http://newio.cc", nil),
			map[string]string{
				"x-amz-bucket-object-lock-enabled": "true",
				// 没有的字段会被忽略
				"x-amz-meta-zero": "x-amz-meta-0",
			},
			"",
			&s3.CreateBucketInput{},
			nil,
		},
		{
			httptest.NewRequest("GET", "http://newio.cc", strings.NewReader("ppp")),
			map[string]string{
				"Content-Length": "999",
			},
			"ppp",
			&s3.UploadPartInput{},
			nil,
		},
		{
			httptest.NewRequest("GET", "http://newio.cc", strings.NewReader("ppp")),
			map[string]string{
				"If-Modified-Since": time.Now().Format(http.TimeFormat),
			},
			"",
			&s3.HeadObjectInput{},
			nil,
		},
	}

	params0 := tests[0].r.URL.Query()
	params0.Add("versionId", "vversionId")
	params0.Add("response-expires", time.Now().Format(http.TimeFormat))
	params0.Add("response-content-type", "vresponse-content-type")
	params0.Add("response-content-language", "vresponse-content-language")
	params0.Add("partNumber", "123123")
	tests[0].r.URL.RawQuery = params0.Encode()
	// tests[0].r.URL.RawQuery = tests[0].r.URL.Query().Encode()

	for k, v := range tests {

		for hh, hv := range v.header {
			v.r.Header.Add(hh, hv)
		}

		errName, err := UnmarshalRequest(context.Background(), v.r, v.data)
		fmt.Println("k -->", k, v.r.URL.RawQuery)
		if err != v.want {
			t.Errorf("k: %v, err: %v,want: %v\n", k, err, v.want)
		}

		_ = errName

		// 暂时这样测试
		if v.payload != "" {
			e := v.data.(*s3.UploadPartInput).Body
			body, err := ioutil.ReadAll(e)
			fmt.Println(string(body) == v.payload, err)
		}

		format(v.r, v.data)
	}
}

func TestMarshalResponse(t *testing.T) {

	tests := []struct {
		w    *httptest.ResponseRecorder
		data interface{}
		want error
	}{
		{httptest.NewRecorder(), &s3.UploadPartOutput{
			ETag:                 aws.String("vETag"),
			RequestCharged:       aws.String("vRequestCharged"),
			SSECustomerAlgorithm: aws.String("vSSECustomerAlgorithm"),
			SSECustomerKeyMD5:    aws.String("vSSECustomerKeyMD5"),
			SSEKMSKeyId:          aws.String("vSSEKMSKeyId"),
			ServerSideEncryption: aws.String("vServerSideEncryption"),
		}, nil},
		{httptest.NewRecorder(), &s3.DeleteObjectOutput{
			DeleteMarker:   aws.Bool(true),
			RequestCharged: aws.String("vRequestCharged"),
			VersionId:      aws.String("vVersionId"),
		}, nil},
		// {httptest.NewRecorder(), &s3.DeleteObjectsOutput{
		// 	Deleted: []*s3.DeletedObject{&s3.DeletedObject{
		// 		Key: aws.String("vKey"),
		// 	}},
		// 	Errors: nil,
		// }, nil},
		{httptest.NewRecorder(), &s3.GetBucketAnalyticsConfigurationOutput{
			AnalyticsConfiguration: &s3.AnalyticsConfiguration{
				Id: aws.String("vid"),
				Filter: &s3.AnalyticsFilter{
					Prefix: aws.String("vPrefix"),
					And: &s3.AnalyticsAndOperator{
						Prefix: aws.String("vvPrefix"),
						Tags: []*s3.Tag{&s3.Tag{
							Key:   aws.String("\"vvkey\""),
							Value: aws.String("vvalue"),
						}},
					},
				},
			},
		}, nil},
		{httptest.NewRecorder(), &s3.GetObjectOutput{
			AcceptRanges: aws.String("vAcceptRanges"),
			Body:         ioutil.NopCloser(bytes.NewReader([]byte("123"))),
			CacheControl: aws.String("vCacheControl"),
			DeleteMarker: aws.Bool(false),
			Metadata: map[string]*string{
				"lala":   aws.String("lala"),
				"github": aws.String("gitlab"),
			},
			PartsCount: aws.Int64(99),
		}, nil},
		{httptest.NewRecorder(), &s3.HeadObjectOutput{
			AcceptRanges: aws.String("vAcceptRanges"),
			DeleteMarker: aws.Bool(true),
			LastModified: aws.Time(time.Now()),
			Metadata: map[string]*string{
				"soda":   aws.String("DomaUmaru"),
				"github": aws.String("gitlab"),
			},
			PartsCount: aws.Int64(99),
		}, nil},
	}

	for k, v := range tests {
		err := MarshalResponse(context.Background(), v.w, v.data)
		if err != v.want {
			t.Errorf("k: %v, err: %v,want: %v\n", k, err, v.want)
		}

		format(v.w.Result())
	}
}

func format(r ...interface{}) {
	fmt.Printf("======\n")
	for _, v := range r {
		spew.Dump(v)
	}
	fmt.Printf("======\n")
}
