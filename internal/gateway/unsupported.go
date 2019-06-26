package gateway

import (
	"context"
	"net/http"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/solution9th/S3Adapter/internal/gerror"
)

type GatewayUnsupported struct{}

func (u GatewayUnsupported) CopyObjectWithContext(ctx context.Context, input *s3.CopyObjectInput, opts ...request.Option) (*s3.CopyObjectOutput, *http.Response, error) {
	return nil, nil, gerror.AWSErrUnsupported
}

func (s *GatewayUnsupported) ListObjectsWithContextV2(ctx context.Context, input *s3.ListObjectsV2Input, opts ...request.Option) (*s3.ListObjectsV2Output, *http.Response, error) {
	return nil, nil, gerror.AWSErrUnsupported
}
