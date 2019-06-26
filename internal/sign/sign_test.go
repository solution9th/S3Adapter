package sign

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/solution9th/S3Adapter/internal/gerror"

	"github.com/stretchr/testify/assert"
)

func TestSignature(t *testing.T) {

	// https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/sig-v4-header-based-auth.html
	var (
		ep = "us-east-1"
	)

	tests := []struct {
		ak, sk string
		r      *http.Request
		h      map[string]string
		sign   string
	}{
		{
			// GET Object https://examplebucket.s3.amazonaws.com/photos/photo1.jpg
			"AKIAIOSFODNN7EXAMPLE",
			"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			httptest.NewRequest("GET", "https://examplebucket.s3.amazonaws.com/test.txt", nil),
			map[string]string{
				"Host":                 "examplebucket.s3.amazonaws.com",
				"x-amz-date":           "20130524T000000Z",
				"Range":                "bytes=0-9",
				"x-amz-content-sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			},
			"AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request, SignedHeaders=host;range;x-amz-content-sha256;x-amz-date, Signature=f0e8bdb87c964420e857bd35b5d6ed310bd44f0170aba48dd91039c6036bdb41",
		},
		{
			// PUT Object Welcome to Amazon S3.
			"AKIAIOSFODNN7EXAMPLE",
			"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			httptest.NewRequest("PUT", "https://examplebucket.s3.amazonaws.com/test$file.text", bytes.NewBufferString("Welcome to Amazon S3.")),
			map[string]string{
				"Date":                 "Fri, 24 May 2013 00:00:00 GMT",
				"x-amz-date":           "20130524T000000Z",
				"x-amz-storage-class":  "REDUCED_REDUNDANCY",
				"x-amz-content-sha256": "44ce7dd67c959e0d3524ffac1771dfbba87d2b6b4b4e99e42034a8b803f8b072",
			},
			"AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request, SignedHeaders=date;host;x-amz-content-sha256;x-amz-date;x-amz-storage-class, Signature=98ad721746da40c64f1a55b78f14c238d841ea1380cd77a1b5971af0ece108bd",
		},
		{
			// GET Bucket Lifecycle
			"AKIAIOSFODNN7EXAMPLE",
			"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			httptest.NewRequest("GET", "https://examplebucket.s3.amazonaws.com/?lifecycle", nil),
			map[string]string{
				"Host":                 "examplebucket.s3.amazonaws.com",
				"x-amz-date":           "20130524T000000Z",
				"x-amz-content-sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			},
			"AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request, SignedHeaders=host;x-amz-content-sha256;x-amz-date, Signature=fea454ca298b7da1c68078a5d1bdbfbbe0d65c699e0f91ac7a200a0136783543",
		},
		{
			// Get Bucket (List Objects)
			"AKIAIOSFODNN7EXAMPLE",
			"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			httptest.NewRequest("GET", "https://examplebucket.s3.amazonaws.com/?max-keys=2&prefix=J", nil),
			map[string]string{
				"Host":                 "examplebucket.s3.amazonaws.com",
				"x-amz-date":           "20130524T000000Z",
				"x-amz-content-sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			},
			"AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request, SignedHeaders=host;x-amz-content-sha256;x-amz-date, Signature=34b48302e7b5fa45bde8084f4b7868a86f0a534bc59db6670ed5711ef69dc6f7",
		},
		{
			"AKIAIYXNSBM66BIOC3TQ",
			"CxNHhX4qUmbDizUZpmxRjUDzyllNPvZicmwG54t6",
			httptest.NewRequest("GET", "http://dont-delete.s3.newio.cc:9091/?list-type=2", nil),
			map[string]string{
				"Host":                 "dont-delete.s3.newio.cc:9091",
				"x-amz-date":           "20190524T072511Z",
				"x-amz-content-sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			},
			"AWS4-HMAC-SHA256 Credential=AKIAIYXNSBM66BIOC3TQ/20190524/us-east-1/s3/aws4_request, SignedHeaders=host;x-amz-content-sha256;x-amz-date, Signature=73b9cfe2b43b4d7474317db59f3fc9896105118b7e10b0ab8c090c1a2255480c",
		},
	}

	for k, v := range tests {
		// if k != 0 {
		// 	continue
		// }

		for hh, hv := range v.h {
			v.r.Header.Set(hh, hv)
		}

		s := NewSignV4(v.ak, v.sk, ep, v.r)
		now, _ := time.Parse(iso8601Format, v.h["x-amz-date"])

		sign, errCode := s.Signature(now)
		assert.Equal(t, gerror.ErrNone, errCode, "k: %v", k)

		if sign != v.sign {
			t.Errorf("k: %v, got: %v, want: %v", k, sign, v.sign)
		} else {
			ss := NewSignV4(v.ak, v.sk, ep, v.r)
			b := ss.Verify(now, sign)
			t.Logf("%v sign:%s", b, sign)
		}
	}

}

func TestVerify(t *testing.T) {

	tests := []struct {
		ak, sk, ep string
		r          *http.Request
		h          map[string]string
	}{
		{
			"AKIAIYXNSBM66BIOC3TQ",
			"CxNHhX4qUmbDizUZpmxRjUDzyllNPvZicmwG54t6",
			"us-east-1",
			httptest.NewRequest("GET", "http://dont-delete.s3.newio.cc:9091/?list-type=2", nil),
			map[string]string{
				"Host":                 "dont-delete.s3.newio.cc:9091",
				"User-Agent":           "aws-sdk-go/1.19.22 (go1.12.5; darwin; amd64)",
				"X-Amz-Content-Sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				"X-Amz-Date":           "20190524T081720Z",
				"Accept-Encoding":      "gzip",
				"Authorization":        "AWS4-HMAC-SHA256 Credential=AKIAIYXNSBM66BIOC3TQ/20190524/us-east-1/s3/aws4_request, SignedHeaders=host;x-amz-content-sha256;x-amz-date, Signature=192bfb73ff304f82bb73cce76dfffb44fe6478f50c8572888cb262c547210127",
			},
		},
		{
			"AKIAIYXNSBM66BIOC3TQ",
			"CxNHhX4qUmbDizUZpmxRjUDzyllNPvZicmwG54t6",
			"us-east-1",
			httptest.NewRequest("HEAD", "/TestS3TestHeadBucket", nil),
			map[string]string{
				"Host":                 "s3.amazonaws.com",
				"X-Amz-Content-Sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				"X-Amz-Date":           "20190605T031650Z",
				"Authorization":        "AWS4-HMAC-SHA256 Credential=AKIAIYXNSBM66BIOC3TQ/20190605/us-east-1/s3/aws4_request, SignedHeaders=host;x-amz-content-sha256;x-amz-date, Signature=abf7104560707a60923dd38d1bbeef8e5043820eda2012accbce017baf1955bd",
			},
		},
	}

	for k, v := range tests {

		for th, tv := range v.h {
			v.r.Header.Add(th, tv)
		}
		a := NewSignV4(v.ak, v.sk, v.ep, v.r)

		now, _ := time.Parse(iso8601Format, v.h["X-Amz-Date"])
		// a.debug = true
		b := a.Verify(now, v.r.Header.Get("Authorization"))

		assert.Equal(t, gerror.ErrNone, b, "k: %v", k)
	}

}

func TestGenSign(t *testing.T) {
	ak := "akak"
	sk := "sksk"
	ep := "beijing"
	now := time.Now().UTC().Format(http.TimeFormat)
	// fmt.Println("now", now)
	r := httptest.NewRequest("DELETE", "http://s3.newio.cc/", nil)
	r.Header.Add("Host", "s3.newio.cc")
	r.Header.Add("Date", now)

	s := NewSignV4(ak, sk, ep, r)
	sign, err := s.Signature(time.Now())

	assert.Equal(t, gerror.ErrNone, err)

	t.Log("sign:", sign)
}
