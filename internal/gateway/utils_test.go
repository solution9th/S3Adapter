package gateway

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestCheckName(t *testing.T) {

	tests := []struct {
		input  interface{}
		fields []string
		want   bool
	}{
		{&s3.CreateBucketInput{
			Bucket: aws.String("123"),
		}, nil, true},
		{&s3.CreateBucketInput{
			Bucket: aws.String("123"),
			ACL:    aws.String("fff"),
		}, []string{"acl"}, true},
		{&s3.CreateBucketInput{
			Bucket: aws.String("123"),
		}, []string{"acl"}, false},
		{&s3.PutObjectInput{
			Bucket: aws.String("123"),
			Key:    aws.String("123"),
		}, nil, true},
		{&s3.CreateBucketInput{
			Bucket: nil,
		}, nil, false},
		{nil, nil, false},
		{s3.CreateBucketInput{}, nil, false},
	}

	for k, v := range tests {
		got := CheckName(v.input, v.fields...)
		if got != v.want {
			t.Errorf("k: %v,got: %v, want: %v", k, got, v.want)
		}
	}
}
