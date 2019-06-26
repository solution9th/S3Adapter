package gateway

import (
	"context"
	"net/http"

	"github.com/solution9th/S3Adapter/internal/auth"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Gateway interface {
	Name() string

	Production() bool

	NewS3Protocol(creds auth.Credentials, region string, isDebug bool) (S3Protocol, error)
}

type S3Protocol interface {

	// =================
	// Bucket operations
	// =================

	CreateBucketWithContext(ctx context.Context, input *s3.CreateBucketInput, opts ...request.Option) (*s3.CreateBucketOutput, *http.Response, error)

	HeadBucketWithContext(ctx context.Context, input *s3.HeadBucketInput, opts ...request.Option) (*s3.HeadBucketOutput, *http.Response, error)

	ListBucketsWithContext(ctx context.Context, input *s3.ListBucketsInput, opts ...request.Option) (*s3.ListBucketsOutput, *http.Response, error)

	ListObjectsWithContext(ctx context.Context, input *s3.ListObjectsInput, opts ...request.Option) (*s3.ListObjectsOutput, *http.Response, error)

	ListObjectsWithContextV2(ctx context.Context, input *s3.ListObjectsV2Input, opts ...request.Option) (*s3.ListObjectsV2Output, *http.Response, error)

	DeleteBucketWithContext(ctx context.Context, input *s3.DeleteBucketInput, opts ...request.Option) (*s3.DeleteBucketOutput, *http.Response, error)

	// =================
	// Object operations
	// =================

	PutObjectWithContext(ctx context.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, *http.Response, error)

	HeadObjectWithContext(ctx context.Context, input *s3.HeadObjectInput, opts ...request.Option) (*s3.HeadObjectOutput, *http.Response, error)

	GetObjectWithContext(ctx context.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, *http.Response, error)

	DeleteObjectWithContext(ctx context.Context, input *s3.DeleteObjectInput, opts ...request.Option) (*s3.DeleteObjectOutput, *http.Response, error)

	CopyObjectWithContext(ctx context.Context, input *s3.CopyObjectInput, opts ...request.Option) (*s3.CopyObjectOutput, *http.Response, error)

	// Multipart operations

	// ACL operations

	// Policy operations
}
