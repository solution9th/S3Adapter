package app

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestFormatWriteXML(t *testing.T) {

	w := httptest.NewRecorder()

	input := &s3.ListObjectsV2Output{
		IsTruncated: aws.Bool(true),
	}

	formatWriteXML(w, 200, "name", input, true)

	fmt.Println(w.Header())
	fmt.Println("===")

	body, err := ioutil.ReadAll(w.Result().Body)
	fmt.Println(string(body), err)
}

func TestWriteErrorResponseXML(t *testing.T) {

	ctx := context.Background()

	w := httptest.NewRecorder()

	e := awserr.NewRequestFailure(awserr.New("IllegalLocationConstraintException", "The unspecified location constraint is incompatible for the region specific endpoint this request was sent to.", nil), 400, "requestsid")

	writeErrorResponseXML(ctx, w, e)

	fmt.Println(w.Header())
	fmt.Println("===")

	body, err := ioutil.ReadAll(w.Result().Body)
	fmt.Println(string(body), err)
}
