package s3

import (
	"context"
	"net/http"

	"github.com/solution9th/S3Adapter/internal/auth"
	"github.com/solution9th/S3Adapter/internal/gateway"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	// Backend s3 backend name
	Backend = "s3"
)

// func init() {
// 	internal.GatewayMap[Backend] = New
// }

// New new Gateway
func New() gateway.Gateway { return &s3gw{} }

type s3gw struct{}

func (s *s3gw) Name() string     { return Backend }
func (s *s3gw) Production() bool { return true }
func (s *s3gw) NewS3Protocol(creds auth.Credentials, region string, isDebug bool) (gateway.S3Protocol, error) {

	cfg := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(creds.AccessKey, creds.SecretKey, creds.SessionToken),
	}

	if isDebug {
		cfg.LogLevel = aws.LogLevel(aws.LogDebugWithHTTPBody)
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}

	// 为了定制化，对 s3 进行包装
	// return s3.New(sess), nil

	return &s3Proto{
		awsClient: s3.New(sess),
	}, nil
}

type s3Proto struct {
	gateway.GatewayUnsupported
	awsClient *s3.S3
}

// =================
// Bucket operations
// =================

func (s *s3Proto) CreateBucketWithContext(ctx context.Context, input *s3.CreateBucketInput, opts ...request.Option) (*s3.CreateBucketOutput, *http.Response, error) {

	var headers http.Header
	r := &http.Response{}

	opts = append(opts, request.WithGetResponseHeaders(&headers))

	output, err := s.awsClient.CreateBucketWithContext(ctx, input, opts...)
	r.Header = headers

	return output, r, err
}

func (s *s3Proto) HeadBucketWithContext(ctx context.Context, input *s3.HeadBucketInput, opts ...request.Option) (*s3.HeadBucketOutput, *http.Response, error) {

	var headers http.Header
	r := &http.Response{}
	opts = append(opts, request.WithGetResponseHeaders(&headers))

	output, err := s.awsClient.HeadBucketWithContext(
		ctx, input, opts...,
	)
	r.Header = headers
	return output, r, err
}

func (s *s3Proto) ListBucketsWithContext(ctx context.Context, input *s3.ListBucketsInput, opts ...request.Option) (*s3.ListBucketsOutput, *http.Response, error) {

	var headers http.Header
	r := &http.Response{}
	opts = append(opts, request.WithGetResponseHeaders(&headers))

	output, err := s.awsClient.ListBucketsWithContext(
		ctx, input, opts...,
	)
	r.Header = headers
	return output, r, err
}

func (s *s3Proto) ListObjectsWithContext(ctx context.Context, input *s3.ListObjectsInput, opts ...request.Option) (*s3.ListObjectsOutput, *http.Response, error) {

	var headers http.Header
	r := &http.Response{}
	opts = append(opts, request.WithGetResponseHeaders(&headers))

	output, err := s.awsClient.ListObjectsWithContext(
		ctx, input, opts...,
	)
	r.Header = headers
	return output, r, err
}

func (s *s3Proto) ListObjectsWithContextV2(ctx context.Context, input *s3.ListObjectsV2Input, opts ...request.Option) (*s3.ListObjectsV2Output, *http.Response, error) {
	var headers http.Header
	r := &http.Response{}
	opts = append(opts, request.WithGetResponseHeaders(&headers))

	output, err := s.awsClient.ListObjectsV2WithContext(ctx, input, opts...)
	r.Header = headers
	return output, r, err
}

func (s *s3Proto) DeleteBucketWithContext(ctx context.Context, input *s3.DeleteBucketInput, opts ...request.Option) (*s3.DeleteBucketOutput, *http.Response, error) {
	var headers http.Header
	r := &http.Response{}
	opts = append(opts, request.WithGetResponseHeaders(&headers))

	output, err := s.awsClient.DeleteBucketWithContext(
		ctx, input, opts...,
	)
	r.Header = headers
	return output, r, err
}

// =================
// Object operations
// =================

func (s *s3Proto) PutObjectWithContext(ctx context.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, *http.Response, error) {
	var headers http.Header
	r := &http.Response{}
	opts = append(opts, request.WithGetResponseHeaders(&headers))

	output, err := s.awsClient.PutObjectWithContext(
		ctx, input, opts...,
	)
	r.Header = headers
	return output, r, err
}

func (s *s3Proto) HeadObjectWithContext(ctx context.Context, input *s3.HeadObjectInput, opts ...request.Option) (*s3.HeadObjectOutput, *http.Response, error) {

	var headers http.Header
	r := &http.Response{}
	opts = append(opts, request.WithGetResponseHeaders(&headers))

	output, err := s.awsClient.HeadObjectWithContext(
		ctx, input, opts...,
	)
	r.Header = headers
	return output, r, err
}

func (s *s3Proto) GetObjectWithContext(ctx context.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, *http.Response, error) {

	var headers http.Header
	r := &http.Response{}
	opts = append(opts, request.WithGetResponseHeaders(&headers))

	output, err := s.awsClient.GetObjectWithContext(
		ctx, input, opts...,
	)
	r.Header = headers
	return output, r, err
}

func (s *s3Proto) DeleteObjectWithContext(ctx context.Context, input *s3.DeleteObjectInput, opts ...request.Option) (*s3.DeleteObjectOutput, *http.Response, error) {

	var headers http.Header
	r := &http.Response{}
	opts = append(opts, request.WithGetResponseHeaders(&headers))

	output, err := s.awsClient.DeleteObjectWithContext(
		ctx, input, opts...,
	)
	r.Header = headers
	return output, r, err
}

func (s *s3Proto) CopyObjectWithContext(ctx context.Context, input *s3.CopyObjectInput, opts ...request.Option) (*s3.CopyObjectOutput, *http.Response, error) {

	var headers http.Header
	r := &http.Response{}
	opts = append(opts, request.WithGetResponseHeaders(&headers))

	output, err := s.awsClient.CopyObjectWithContext(
		ctx, input, opts...,
	)
	r.Header = headers
	return output, r, err
}
