package cos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/solution9th/S3Adapter/internal/auth"
	"github.com/solution9th/S3Adapter/internal/gateway"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dustin/go-humanize"
	"github.com/haozibi/zlog"
	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/tencentyun/cos-go-sdk-v5/debug"
)

var (
	defaultServiceBaseURL = "http://service.cos.myqcloud.com"
)

const (
	cosS3MinPartSize        = 5 * humanize.MiByte
	cosMaxParts             = 1000
	cosMaxKeys              = 1000
	HTTPHeaderCosMetaPrefix = "X-Cos-Meta-"
	HTTPHeaderS3MetaPrefix  = "X-Amz-Meta-"
)

const (
	// Backend cos backend name
	Backend = "cos"
)

// New new Gateway
func New() gateway.Gateway { return &cosgw{} }

type cosgw struct{}

func (s *cosgw) Name() string     { return Backend }
func (s *cosgw) Production() bool { return true }
func (s *cosgw) NewS3Protocol(creds auth.Credentials, region string, isDebug bool) (gateway.S3Protocol, error) {

	cfg := &cos.AuthorizationTransport{
		SecretID:     creds.AccessKey,
		SecretKey:    creds.SecretKey,
		SessionToken: creds.SessionToken,
	}

	if isDebug {
		cfg.Transport = &debug.DebugRequestTransport{
			RequestHeader:  true,
			RequestBody:    true,
			ResponseHeader: true,
			ResponseBody:   true,
		}
	}

	c := cos.NewClient(nil, &http.Client{
		Transport: cfg,
	})

	return &cosProto{
		cosClient: c,
		region:    region,
		appid:     creds.AccessKey,
		cosURI:    "https://%v.cos.%s.myqcloud.com",
	}, nil
}

type cosProto struct {
	gateway.GatewayUnsupported
	cosClient *cos.Client
	region    string
	appid     string
	cosURI    string
}

// =================
// Bucket operations
// =================

func (s *cosProto) CreateBucketWithContext(ctx context.Context, input *s3.CreateBucketInput, opts ...request.Option) (*s3.CreateBucketOutput, *http.Response, error) {

	cosInput := &cos.BucketPutOptions{}

	if input.ACL != nil {
		cosInput.XCosACL = aws.StringValue(input.ACL)
	}

	if input.GrantRead != nil {
		cosInput.XCosGrantRead = aws.StringValue(input.GrantRead)
	}

	if input.GrantWrite != nil {
		cosInput.XCosGrantWrite = aws.StringValue(input.GrantWrite)
	}

	if input.GrantFullControl != nil {
		cosInput.XCosGrantFullControl = aws.StringValue(input.GrantFullControl)
	}

	// u, _ := url.Parse("http://<BucketName-APPID>.cos.<region>.myqcloud.com")

	u, _ := url.Parse(fmt.Sprintf(s.cosURI, aws.StringValue(input.Bucket), s.region))
	s.cosClient.BaseURL.BucketURL = u

	resp, err := s.cosClient.Bucket.Put(ctx, cosInput)
	if err != nil {
		zlog.ZError().Str("Method", "CreateBucketWithContext").Str("Bucket", aws.StringValue(input.Bucket)).Msg("[COS] error:" + err.Error())
		return nil, getErrResponse(err), toS3Err(err)
	}

	resp.Response.Header.Set(responseRequestIDKey, resp.Response.Header.Get("x-cos-request-id"))
	resp.Response.Header.Set(responseAMZIDKey, resp.Response.Header.Get("x-cos-trace-id"))

	return &s3.CreateBucketOutput{
		Location: aws.String(s.region),
	}, resp.Response, nil
}

func (s *cosProto) HeadBucketWithContext(ctx context.Context, input *s3.HeadBucketInput, opts ...request.Option) (*s3.HeadBucketOutput, *http.Response, error) {

	u, _ := url.Parse(fmt.Sprintf(s.cosURI, aws.StringValue(input.Bucket), s.region))
	s.cosClient.BaseURL.BucketURL = u

	resp, err := s.cosClient.Bucket.Head(ctx)
	if err != nil {
		zlog.ZError().Str("Method", "HeadBucketWithContext").Str("Bucket", aws.StringValue(input.Bucket)).Msg("[COS] error:" + err.Error())
		return nil, getErrResponse(err), toS3Err(err)
	}

	resp.Response.Header.Set(responseRequestIDKey, resp.Response.Header.Get("x-cos-request-id"))
	resp.Response.Header.Set(responseAMZIDKey, resp.Response.Header.Get("x-cos-trace-id"))

	return &s3.HeadBucketOutput{}, resp.Response, nil
}

func (s *cosProto) DeleteBucketWithContext(ctx context.Context, input *s3.DeleteBucketInput, opts ...request.Option) (*s3.DeleteBucketOutput, *http.Response, error) {

	u, _ := url.Parse(fmt.Sprintf(s.cosURI, aws.StringValue(input.Bucket), s.region))
	s.cosClient.BaseURL.BucketURL = u

	resp, err := s.cosClient.Bucket.Delete(ctx)
	if err != nil {
		zlog.ZError().Str("Method", "DeleteBucketWithContext").Str("Bucket", aws.StringValue(input.Bucket)).Msg("[COS] error:" + err.Error())
		return nil, getErrResponse(err), toS3Err(err)
	}
	resp.Response.Header.Set(responseRequestIDKey, resp.Response.Header.Get("x-cos-request-id"))
	resp.Response.Header.Set(responseAMZIDKey, resp.Response.Header.Get("x-cos-trace-id"))
	return &s3.DeleteBucketOutput{}, resp.Response, nil
}

func (s *cosProto) ListObjectsWithContext(ctx context.Context, input *s3.ListObjectsInput, opts ...request.Option) (*s3.ListObjectsOutput, *http.Response, error) {

	u, _ := url.Parse(fmt.Sprintf(s.cosURI, aws.StringValue(input.Bucket), s.region))
	s.cosClient.BaseURL.BucketURL = u

	cosInput := &cos.BucketGetOptions{}
	if input.Delimiter != nil {
		cosInput.Delimiter = aws.StringValue(input.Delimiter)
	}

	if input.EncodingType != nil {
		cosInput.EncodingType = aws.StringValue(input.EncodingType)
	}

	if input.Marker != nil {
		cosInput.Marker = aws.StringValue(input.Marker)
	}

	if input.Prefix != nil {
		cosInput.Prefix = aws.StringValue(input.Prefix)
	}

	if input.MaxKeys != nil {
		cosInput.MaxKeys = int(aws.Int64Value(input.MaxKeys))
	}

	cosOutput, resp, err := s.cosClient.Bucket.Get(ctx, cosInput)
	if err != nil {
		zlog.ZError().Str("Method", "ListObjectsWithContext").Str("Bucket", aws.StringValue(input.Bucket)).Msg("[COS] error:" + err.Error())
		resp.Response.Header.Set(responseRequestIDKey, resp.Response.Header.Get("x-cos-request-id"))
		resp.Response.Header.Set(responseAMZIDKey, resp.Response.Header.Get("x-cos-trace-id"))
		return nil, resp.Response, toS3Err(err)
	}

	output := &s3.ListObjectsOutput{
		Name:         aws.String(cosOutput.Name),
		Prefix:       aws.String(cosOutput.Prefix),
		Marker:       aws.String(cosOutput.Marker),
		NextMarker:   aws.String(cosOutput.NextMarker),
		Delimiter:    aws.String(cosOutput.Delimiter),
		MaxKeys:      aws.Int64(int64(cosOutput.MaxKeys)),
		IsTruncated:  aws.Bool(cosOutput.IsTruncated),
		EncodingType: aws.String(cosOutput.EncodingType),
	}

	output.CommonPrefixes = make([]*s3.CommonPrefix, len(cosOutput.CommonPrefixes))
	for i := 0; i < len(cosOutput.CommonPrefixes); i++ {
		output.CommonPrefixes[i].Prefix = aws.String(cosOutput.CommonPrefixes[i])
	}

	output.Contents = make([]*s3.Object, len(cosOutput.Contents))
	for i := 0; i < len(cosOutput.Contents); i++ {

		// 2019-06-13T08:30:15.000Z
		lmt, err := time.Parse(time.RFC3339, cosOutput.Contents[i].LastModified)
		if err != nil {
			zlog.ZError().Msg(err.Error())
		}

		output.Contents[i] = &s3.Object{
			ETag:         aws.String(cosOutput.Contents[i].ETag),
			Key:          aws.String(cosOutput.Contents[i].Key),
			LastModified: aws.Time(lmt),
			Size:         aws.Int64(int64(cosOutput.Contents[i].Size)),
			StorageClass: aws.String(cosOutput.Contents[i].StorageClass),
		}

		output.Contents[i].Owner = &s3.Owner{
			DisplayName: aws.String(cosOutput.Contents[i].Owner.DisplayName),
			ID:          aws.String(cosOutput.Contents[i].Owner.ID),
		}
	}

	resp.Response.Header.Set(responseRequestIDKey, resp.Response.Header.Get("x-cos-request-id"))
	resp.Response.Header.Set(responseAMZIDKey, resp.Response.Header.Get("x-cos-trace-id"))

	return output, resp.Response, nil
}

func (s *cosProto) ListBucketsWithContext(ctx context.Context, input *s3.ListBucketsInput, opts ...request.Option) (*s3.ListBucketsOutput, *http.Response, error) {

	s.cosClient.BaseURL.BucketURL, _ = url.Parse(defaultServiceBaseURL)

	cosOutput, resp, err := s.cosClient.Service.Get(ctx)
	if err != nil {
		zlog.ZError().Str("Method", "ListBucketsWithContext").Msg("[COS] error:" + err.Error())
		return nil, getErrResponse(err), toS3Err(err)
	}

	output := &s3.ListBucketsOutput{
		Owner: &s3.Owner{
			DisplayName: aws.String(cosOutput.Owner.DisplayName),
			ID:          aws.String(cosOutput.Owner.ID),
		},
	}

	output.Buckets = make([]*s3.Bucket, len(cosOutput.Buckets))

	for i := 0; i < len(cosOutput.Buckets); i++ {
		lmt, _ := time.Parse(time.RFC3339, cosOutput.Buckets[i].CreationDate)

		output.Buckets[i] = &s3.Bucket{
			Name:         aws.String(cosOutput.Buckets[i].Name),
			CreationDate: aws.Time(lmt),
		}
	}

	resp.Response.Header.Set(responseRequestIDKey, resp.Response.Header.Get("x-cos-request-id"))
	resp.Response.Header.Set(responseAMZIDKey, resp.Response.Header.Get("x-cos-trace-id"))

	return output, resp.Response, nil
}

// =================
// Object operations
// =================

func s3HeaderToCosMeta(s3Metadata map[string]*string) *http.Header {
	header := make(http.Header)

	for k, v := range s3Metadata {
		k = http.CanonicalHeaderKey(k)
		switch {
		case strings.HasPrefix(k, HTTPHeaderS3MetaPrefix):
			metaKey := k[len(HTTPHeaderS3MetaPrefix):]
			if strings.Contains(metaKey, "_") {
				zlog.ZError().Str("method", "s3HeaderToCosMeta").Msg(k)
				return nil
			}
			header.Set(HTTPHeaderCosMetaPrefix+metaKey, *v)
		}
	}
	return &header
}

func cosHeaderToS3Header(header http.Header) map[string]*string {

	s3Metadata := make(map[string]*string)
	fmt.Println(header)
	for k := range header {
		fmt.Println(k)
		k = http.CanonicalHeaderKey(k)
		fmt.Println(k)
		switch {
		case strings.HasPrefix(k, HTTPHeaderCosMetaPrefix):
			// Add amazon s3 meta prefix
			metaKey := k[len(HTTPHeaderCosMetaPrefix):]
			metaKey = HTTPHeaderS3MetaPrefix + metaKey
			metaKey = http.CanonicalHeaderKey(metaKey)
			s3Metadata[metaKey] = aws.String(header.Get(k))
		}
	}

	return s3Metadata
}

func (s *cosProto) GetObjectWithContext(ctx context.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, *http.Response, error) {
	bucket := *input.Bucket
	object := *input.Key

	u, _ := url.Parse(fmt.Sprintf(s.cosURI, aws.StringValue(input.Bucket), s.region))
	s.cosClient.BaseURL.BucketURL = u

	opt := &cos.ObjectGetOptions{}
	if input.ResponseContentType != nil {
		opt.ResponseContentType = aws.StringValue(input.ResponseContentType)
	}

	if input.ResponseCacheControl != nil {
		opt.ResponseCacheControl = aws.StringValue(input.ResponseCacheControl)
	}

	if input.ResponseContentLanguage != nil {
		opt.ResponseContentLanguage = aws.StringValue(input.ResponseContentLanguage)
	}

	if input.ResponseContentDisposition != nil {
		opt.ResponseContentDisposition = aws.StringValue(input.ResponseContentDisposition)
	}

	if input.ResponseContentEncoding != nil {
		opt.ResponseContentEncoding = aws.StringValue(input.ResponseContentEncoding)
	}

	if input.ResponseExpires != nil {
		opt.ResponseExpires = input.ResponseExpires.Format(http.TimeFormat)
	}

	if input.IfModifiedSince != nil {
		opt.IfModifiedSince = input.IfModifiedSince.Format(http.TimeFormat)
	}

	if input.Range != nil {
		opt.Range = aws.StringValue(input.Range)
	}

	//opt可选，无特殊设置可设为nil
	resp, err := s.cosClient.Object.Get(ctx, object, opt)
	resp.Response.Header.Set(responseRequestIDKey, resp.Response.Header.Get("x-cos-request-id"))
	resp.Response.Header.Set(responseAMZIDKey, resp.Response.Header.Get("x-cos-trace-id"))
	if err != nil {
		zlog.ZError().Str("method", "Object.Get").Str("bucket", bucket).Msg(err.Error())
		return nil, resp.Response, toS3Err(err)
	}

	header := resp.Header
	//data, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	panic(err)
	//}
	//err = ioutil.WriteFile("test.png", data, 0644)
	//if err != nil {
	//	panic(err)
	//}

	return &s3.GetObjectOutput{
		Body:                 resp.Body,
		VersionId:            awsString(header.Get("x-cos-version-id")),
		StorageClass:         awsString(header.Get("x-cos-storage-class")),
		Metadata:             cosHeaderToS3Header(header),
		ServerSideEncryption: awsString(header.Get("x-cos-server-side-encryption")),
	}, resp.Response, nil
}

func (s *cosProto) HeadObjectWithContext(ctx context.Context, input *s3.HeadObjectInput, opts ...request.Option) (*s3.HeadObjectOutput, *http.Response, error) {
	bucket := *input.Bucket
	object := *input.Key

	u, _ := url.Parse(fmt.Sprintf(s.cosURI, aws.StringValue(input.Bucket), s.region))
	s.cosClient.BaseURL.BucketURL = u

	opt := &cos.ObjectHeadOptions{}
	if input.IfModifiedSince != nil {
		opt.IfModifiedSince = aws.TimeValue(input.IfModifiedSince).Format(http.TimeFormat)
	}
	resp, err := s.cosClient.Object.Head(ctx, object, opt)
	resp.Response.Header.Set(responseRequestIDKey, resp.Response.Header.Get("x-cos-request-id"))
	resp.Response.Header.Set(responseAMZIDKey, resp.Response.Header.Get("x-cos-trace-id"))
	if err != nil {
		fmt.Println(err)
		zlog.ZError().Str("method", "Object.Get").Str("bucket", bucket).Str("object", object).Msg(err.Error())
		return nil, resp.Response, toS3Err(err)
	}

	header := resp.Header

	// Build S3 metadata from COS metadata
	meta := cosHeaderToS3Header(header)

	// modTime, err := time.Parse(time.RFC3339, header.Get("Last-Modified"))
	modTime, err := time.Parse(http.TimeFormat, header.Get("Last-Modified"))
	if err != nil {
		return nil, resp.Response, toS3Err(err)
	}
	size, err := strconv.ParseInt(header.Get("Content-Length"), 10, 64)
	if err != nil {
		return nil, resp.Response, toS3Err(err)
	}

	return &s3.HeadObjectOutput{
		LastModified:    aws.Time(modTime),
		ETag:            awsString(header.Get("ETag")),
		Metadata:        meta,
		ContentType:     awsString(header.Get("Content-Type")),
		ContentLength:   &size,
		ContentEncoding: awsString(header.Get("Content-Encoding")),
	}, resp.Response, nil
}

func awsString(v string) *string {
	if v != "" {
		return &v
	}
	return nil
}

func (s *cosProto) PutObjectWithContext(ctx context.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, *http.Response, error) {
	bucket := *input.Bucket
	object := *input.Key

	u, _ := url.Parse(fmt.Sprintf(s.cosURI, aws.StringValue(input.Bucket), s.region))
	s.cosClient.BaseURL.BucketURL = u

	// Build COS metadata
	opt := &cos.ObjectPutOptions{
		ACLHeaderOptions: &cos.ACLHeaderOptions{
			//XCosGrantWrite:       nil,
		},
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			//Expect:             nil,
			//XCosContentSHA1:    nil,
			XCosMetaXXX: s3HeaderToCosMeta(input.Metadata),
		},
	}
	if input.GrantRead != nil {
		opt.XCosGrantRead = aws.StringValue(input.GrantRead)
	}

	if input.GrantFullControl != nil {
		opt.XCosGrantFullControl = aws.StringValue(input.GrantFullControl)
	}

	if input.ACL != nil {
		opt.XCosACL = aws.StringValue(input.ACL)
	}

	if input.CacheControl != nil {
		opt.CacheControl = aws.StringValue(input.CacheControl)
	}

	if input.ContentDisposition != nil {
		opt.ContentDisposition = aws.StringValue(input.ContentDisposition)
	}

	if input.ContentEncoding != nil {
		opt.ContentEncoding = aws.StringValue(input.ContentEncoding)
	}

	if input.ContentType != nil {
		opt.ContentType = aws.StringValue(input.ContentType)
	}

	if input.ContentLength != nil {
		opt.ContentLength = int(aws.Int64Value(input.ContentLength))
	}

	if input.Expires != nil {
		opt.Expires = input.Expires.Format(http.TimeFormat)
	}

	if input.StorageClass != nil {
		opt.XCosStorageClass = aws.StringValue(input.StorageClass)
	}

	resp, err := s.cosClient.Object.Put(ctx, object, input.Body, opt)
	resp.Response.Header.Set(responseRequestIDKey, resp.Response.Header.Get("x-cos-request-id"))
	resp.Response.Header.Set(responseAMZIDKey, resp.Response.Header.Get("x-cos-trace-id"))
	if err != nil {
		zlog.ZError().Str("method", "PutObject").Str("bucket", bucket).Str("object", object).Msg(err.Error())
		return nil, resp.Response, toS3Err(err)
	}

	header := resp.Header
	return &s3.PutObjectOutput{
		ETag:                 awsString(header.Get("Etag")),
		Expiration:           nil,
		RequestCharged:       nil,
		SSECustomerAlgorithm: nil,
		SSECustomerKeyMD5:    nil,
		SSEKMSKeyId:          nil,
		ServerSideEncryption: awsString(header.Get("x-cos-server-side-encryption")),
		VersionId:            awsString(header.Get("x-cos-version-id")),
	}, resp.Response, nil
}

func (s *cosProto) DeleteObjectWithContext(ctx context.Context, input *s3.DeleteObjectInput, opts ...request.Option) (*s3.DeleteObjectOutput, *http.Response, error) {
	bucket := *input.Bucket
	object := *input.Key

	u, _ := url.Parse(fmt.Sprintf(s.cosURI, aws.StringValue(input.Bucket), s.region))
	s.cosClient.BaseURL.BucketURL = u

	resp, err := s.cosClient.Object.Delete(ctx, object)
	resp.Response.Header.Set(responseRequestIDKey, resp.Response.Header.Get("x-cos-request-id"))
	resp.Response.Header.Set(responseAMZIDKey, resp.Response.Header.Get("x-cos-trace-id"))
	if err != nil {
		zlog.ZError().Str("method", "DeleteObject").Str("bucket", bucket).Str("object", object).Msg(err.Error())
		return nil, resp.Response, toS3Err(err)
	}
	return &s3.DeleteObjectOutput{
		DeleteMarker:   nil,
		RequestCharged: nil,
		VersionId:      nil,
	}, resp.Response, nil
}

func (s *cosProto) CopyObjectWithContext(ctx context.Context, input *s3.CopyObjectInput, opts ...request.Option) (*s3.CopyObjectOutput, *http.Response, error) {
	bucket := *input.Bucket
	object := *input.Key

	u, _ := url.Parse(fmt.Sprintf(s.cosURI, aws.StringValue(input.Bucket), s.region))
	s.cosClient.BaseURL.BucketURL = u

	opt := &cos.ObjectCopyOptions{
		ObjectCopyHeaderOptions: &cos.ObjectCopyHeaderOptions{
			XCosMetaXXX: s3HeaderToCosMeta(input.Metadata),
		},
		ACLHeaderOptions: &cos.ACLHeaderOptions{
			//XCosGrantWrite: nil,
		},
	}

	if input.ACL != nil {
		opt.XCosACL = aws.StringValue(input.ACL)
	}

	if input.GrantFullControl != nil {
		opt.XCosGrantFullControl = aws.StringValue(input.GrantFullControl)
	}

	if input.GrantRead != nil {
		opt.XCosGrantRead = aws.StringValue(input.GrantRead)
	}

	if input.StorageClass != nil {
		opt.XCosStorageClass = aws.StringValue(input.StorageClass)
	}

	if input.CopySource != nil {
		opt.XCosCopySource = aws.StringValue(input.CopySource)
	}

	if input.CopySourceIfMatch != nil {
		opt.XCosCopySourceIfMatch = aws.StringValue(input.CopySourceIfMatch)
	}

	if input.CopySourceIfModifiedSince != nil {
		opt.XCosCopySourceIfModifiedSince = input.CopySourceIfModifiedSince.Format(http.TimeFormat)
	}

	if input.CopySourceIfNoneMatch != nil {
		opt.XCosCopySourceIfNoneMatch = aws.StringValue(input.CopySourceIfNoneMatch)
	}

	if input.CopySourceIfUnmodifiedSince != nil {
		opt.XCosCopySourceIfUnmodifiedSince = input.CopySourceIfUnmodifiedSince.Format(http.TimeFormat)
	}

	if input.MetadataDirective != nil {
		opt.XCosMetadataDirective = aws.StringValue(input.MetadataDirective)
	}

	res, resp, err := s.cosClient.Object.Copy(ctx, object, opt.XCosCopySource, opt)
	if err != nil {
		zlog.ZError().Str("method", "CopyObject").Str("bucket", bucket).Str("object", object).Msg(err.Error())
		return nil, resp.Response, toS3Err(err)
	}
	header := resp.Header
	lastMod, err := time.Parse(time.RFC3339, res.LastModified)
	if err != nil {
		zlog.ZError().Str("method", "CopyObject").Str("bucket", bucket).Str("object", object).Msg(err.Error())
		return nil, resp.Response, toS3Err(err)
	}
	return &s3.CopyObjectOutput{
		CopyObjectResult: &s3.CopyObjectResult{
			ETag:         aws.String(res.ETag),
			LastModified: &lastMod,
		},
		ServerSideEncryption: awsString(header.Get("x-cos-server-side-encryption")),
		VersionId:            awsString(header.Get("x-cos-version-id")),
	}, resp.Response, nil
}
