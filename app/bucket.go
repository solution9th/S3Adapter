package app

import (
	"encoding/xml"
	"io"
	"net/http"
	"strconv"

	"github.com/solution9th/S3Adapter/internal/gerror"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	"github.com/haozibi/zlog"
)

const (
	maxObjectList = 1000
)

// PutApplication 创建应用，换取 key,(暂时没有签名)
//
// 独有方法，新获得的 key，主要是为了区分 后端存储引擎 和 具体的 key
// 请求:
// 	<CreateApplicationConfiguration>
// 		<AccessKey></AccessKey>
// 		<SecretKey></SecretKey>
// 		<Engine></Engine>
// 		<Region></Region>
// 		<AppName></AppName>
// 		<AppRemark></AppRemark>
// 	</CreateApplicationConfiguration>
//
// 响应:
// 	<?xml version="1.0" encoding="UTF-8"?>
// 	<CreateApplicationResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
// 		<AccessKey></AccessKey>
// 		<SecretKey></SecretKey>
// 	</CreateApplicationResult>
func (a *API) PutApplication(w http.ResponseWriter, r *http.Request) {

	ctx := newContext(w, r, "PutApplication")

	zlog.ZDebug().Str("Method", "PutApplication").Msg("[debug]")

	var configLocation CreateApplicationConfiguration
	if r.Body != nil {
		err := xml.NewDecoder(r.Body).Decode(&configLocation)
		if err != nil {
			zlog.ZError().Str("Method", "XMLDecode").Msg(err.Error())
			writeErrorResponseXML(ctx, w,
				gerror.GetError(gerror.ErrMalformedXML, err))
			return
		}
	}

	oak, osk, errCode := a.saveInfo(configLocation)
	if errCode != gerror.ErrNone {
		writeErrorResponseXML(ctx, w,
			gerror.GetError(errCode, nil))
		return
	}

	type CreateApplicationResult struct {
		AccessKey string `xml:"AccessKey"`
		SecretKey string `xml:"SecretKey"`
	}

	var tr CreateApplicationResult

	tr.AccessKey = oak
	tr.SecretKey = osk

	formatWriteXML(w, http.StatusOK, "", tr, true)
	return
}

// DeleteApplication 删除应用，通过 AWS Sign V4 验证签名
func (a *API) DeleteApplication(w http.ResponseWriter, r *http.Request) {

	ctx := newContext(w, r, "DeleteApplication")

	zlog.ZDebug().Str("Method", "DeleteApplication").Msg("[debug]")

	if errCode := a.Auth(r); errCode != gerror.ErrNone {
		if errCode == gerror.ErrRequestTimeTooSkewed {
			writeErrorRequestTimeTooSkewed(ctx, w, r)
			return
		}
		writeErrorResponseXML(ctx, w,
			gerror.GetError(errCode, nil))
		return
	}

	auth := a.getAuthorizationInfo(r)
	// if auth == nil {
	// 	writeErrorResponseXML(ctx, w,
	// 		gerror.GetError(gerror.ErrAllAccessDisabled, nil))
	// 	return
	// }

	err := a.deleteInfo(auth.oak, auth.osk)
	if err != nil {
		writeErrorResponseXML(ctx, w,
			gerror.GetError(gerror.ErrServerNotInitialized, nil))
		return
	}
	writeSuccessResponseHeadersOnly(w)
}

// HeadBucket https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/RESTBucketHEAD.html
//
// This operation is useful to determine if a bucket exists and you have permission to access it.
// The operation returns a 200 OK if the bucket exists and you have permission to access it.
// Otherwise, the operation might return responses such as 404 Not Found and 403 Forbidden.
//
// To use this operation, you must have permissions to perform the s3:ListBucket action.
// The bucket owner has this permission by default and can grant this permission to others.
// For more information about permissions,
// see Permissions Related to Bucket Operations  and Managing Access Permissions to
// Your Amazon S3 Resources in the Amazon Simple Storage Service Developer Guide.
//
// https://docs.aws.amazon.com/AmazonS3/latest/dev/using-with-s3-actions.html#using-with-s3-actions-related-to-buckets
// https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-access-control.html
func (a *API) HeadBucket(w http.ResponseWriter, r *http.Request) {

	ctx := newContext(w, r, "HeadBucket")

	vars := mux.Vars(r)
	bucket := vars["bucket"]

	zlog.ZDebug().Str("Bucket", bucket).Str("Method", "HeadBucket").Msg("[debug]")

	if errCode := a.Auth(r); errCode != gerror.ErrNone {
		if errCode == gerror.ErrRequestTimeTooSkewed {
			writeErrorRequestTimeTooSkewed(ctx, w, r)
			return
		}
		writeErrorResponseXML(ctx, w,
			gerror.GetError(errCode, nil))
		return
	}

	gProto := a.GetGateway(r)
	if gProto == nil {
		writeErrorResponseXML(ctx, w,
			gerror.GetError(gerror.ErrServerNotInitialized, nil))
		return
	}

	// 涉及到 s3:ListBucket，暂时不检查

	_, resp, err := gProto.HeadBucketWithContext(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket)})
	if err != nil {
		writeS3Header(w, resp.Header)
		writeErrorResponseHeadersOnly(w, err)
		return
	}

	writeS3Header(w, resp.Header)
}

// GetBucketV2 https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/v2-RESTBucketGET.html
//
// This implementation of the GET operation returns some or all (up to 1,000) of the objects in a bucket.
// You can use the request parameters as selection criteria to return a subset of the objects in a bucket.
// A 200 OK response can contain valid or invalid XML.
// Make sure to design your application to parse the contents of the response and handle it appropriately.
//
// To use this implementation of the operation, you must have READ access to the bucket.
//
// To use this operation in an AWS Identity and Access Management (IAM) policy, you must have permissions to perform the s3:ListBucket action.
// The bucket owner has this permission by default and can grant this permission to others.
// For more information about permissions, see Permissions Related to Bucket Operations and Managing Access Permissions to
// Your Amazon S3 Resources in the Amazon Simple Storage Service Developer Guide.
func (a *API) GetBucketV2(w http.ResponseWriter, r *http.Request) {

	ctx := newContext(w, r, "GetBucketV2")

	vars := mux.Vars(r)
	bucket := vars["bucket"]

	zlog.ZDebug().Str("Bucket", bucket).Str("Method", "GetBucketV2").Msg("[debug]")

	if errCode := a.Auth(r); errCode != gerror.ErrNone {
		if errCode == gerror.ErrRequestTimeTooSkewed {
			writeErrorRequestTimeTooSkewed(ctx, w, r)
			return
		}
		writeErrorResponseXML(ctx, w,
			gerror.GetError(errCode, nil))
		return
	}

	gProto := a.GetGateway(r)
	if gProto == nil {
		writeErrorResponseXML(ctx, w,
			gerror.GetError(gerror.ErrServerNotInitialized, nil))
		return
	}

	uriValues := r.URL.Query()

	if v, ok := uriValues["continuation-token"]; ok {
		if len(v[0]) == 0 {
			writeErrorResponseXML(ctx, w,
				gerror.GetError(gerror.ErrIncorrectContinuationToken, nil),
				AddArg("continuation-token"),
			)
			return
		}
	}

	maxKeys := maxObjectList
	if uriValues.Get("max-keys") != "" {
		var err error
		if maxKeys, err = strconv.Atoi(uriValues.Get("max-keys")); err != nil {
			writeErrorResponseXML(ctx, w,
				gerror.GetError(gerror.ErrInvalidMaxKeys, err),
				AddArg("max-keys"),
				AddArgValue(uriValues.Get("max-keys")),
			)
			return
		}
	}

	prefix := uriValues.Get("prefix")
	startAfter := uriValues.Get("start-after")
	delimiter := uriValues.Get("delimiter")
	fetchOwner := uriValues.Get("fetch-owner") == "true"
	// token := uriValues.Get("continuation-token")
	encodingType := uriValues.Get("encoding-type")

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		// ContinuationToken: aws.String(token),
		Delimiter:  aws.String(delimiter),
		FetchOwner: aws.Bool(fetchOwner),
		MaxKeys:    aws.Int64(int64(maxKeys)),
		Prefix:     aws.String(prefix),
		StartAfter: aws.String(startAfter),
	}

	if encodingType != "" {
		if encodingType == "url" {
			input.EncodingType = aws.String(encodingType)
		} else {
			writeErrorResponseXML(ctx, w,
				gerror.GetError(gerror.ErrInvalidEncodingMethod, nil),
				AddArg("encoding-type"),
				AddArgValue(encodingType),
			)
			return
		}
	}

	output, resp, err := gProto.ListObjectsWithContextV2(ctx, input)
	if err != nil {
		writeS3Header(w, resp.Header)
		writeErrorResponseXML(ctx, w, err)
		return
	}

	writeS3Header(w, resp.Header)
	formatWriteXML(w, http.StatusOK, "ListBucketResult", output, true)

	return
}

// GetBucketV1 https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/RESTBucketGET.html
//
// This implementation of the GET operation returns some or all (up to 1,000) of the objects in a bucket.
// You can use the request parameters as selection criteria to return a subset of the objects in a bucket.
// A 200 OK response can contain valid or invalid XML.
// Be sure to design your application to parse the contents of the response and handle it appropriately.
func (a *API) GetBucketV1(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(w, r, "GetBucketV1")

	vars := mux.Vars(r)
	bucket := vars["bucket"]

	zlog.ZDebug().Str("Bucket", bucket).Str("Method", "GetBucketV1").Msg("[debug]")

	if errCode := a.Auth(r); errCode != gerror.ErrNone {
		if errCode == gerror.ErrRequestTimeTooSkewed {
			writeErrorRequestTimeTooSkewed(ctx, w, r)
			return
		}
		writeErrorResponseXML(ctx, w,
			gerror.GetError(errCode, nil))
		return
	}

	gProto := a.GetGateway(r)
	if gProto == nil {
		writeErrorResponseXML(ctx, w,
			gerror.GetError(gerror.ErrServerNotInitialized, nil))
		return
	}

	uriValues := r.URL.Query()

	delimiter := uriValues.Get("delimiter")
	encodingType := uriValues.Get("encoding-type")
	marker := uriValues.Get("marker")
	prefix := uriValues.Get("prefix")

	maxKeys := maxObjectList
	if uriValues.Get("max-keys") != "" {
		var err error
		if maxKeys, err = strconv.Atoi(uriValues.Get("max-keys")); err != nil {
			writeErrorResponseXML(ctx, w,
				gerror.GetError(gerror.ErrInvalidMaxKeys, err),
				AddArg("max-keys"),
				AddArgValue(uriValues.Get("max-keys")),
			)
			return
		}
	}

	input := &s3.ListObjectsInput{
		Bucket:       aws.String(bucket),
		Delimiter:    aws.String(delimiter),
		EncodingType: aws.String(encodingType),
		Marker:       aws.String(marker),
		MaxKeys:      aws.Int64(int64(maxKeys)),
		Prefix:       aws.String(prefix),
	}

	if encodingType != "" {
		if encodingType == "url" {
			input.EncodingType = aws.String(encodingType)
		} else {
			writeErrorResponseXML(ctx, w,
				gerror.GetError(gerror.ErrInvalidEncodingMethod, nil),
				AddArg("encoding-type"),
				AddArgValue(encodingType),
			)
			return
		}
	}

	output, resp, err := gProto.ListObjectsWithContext(ctx, input)
	if err != nil {
		writeS3Header(w, resp.Header)
		writeErrorResponseXML(ctx, w, err)
		return
	}

	writeS3Header(w, resp.Header)
	formatWriteXML(w, http.StatusOK, "ListBucketResult", output, true)

	return
}

// PutBucket https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/RESTBucketPUT.html
//
// This implementation of the PUT operation creates a new bucket.
// To create a bucket, you must register with Amazon S3 and have a valid AWS Access Key ID to authenticate requests.
// Anonymous requests are never allowed to create buckets. By creating the bucket, you become the bucket owner.
func (a *API) PutBucket(w http.ResponseWriter, r *http.Request) {

	ctx := newContext(w, r, "PutBucket")

	vars := mux.Vars(r)
	bucket := vars["bucket"]

	zlog.ZDebug().Str("Bucket", bucket).Str("Method", "PutBucket").Msg("[debug]")

	if errCode := a.Auth(r); errCode != gerror.ErrNone {
		if errCode == gerror.ErrRequestTimeTooSkewed {
			writeErrorRequestTimeTooSkewed(ctx, w, r)
			return
		}
		writeErrorResponseXML(ctx, w,
			gerror.GetError(errCode, nil))
		return
	}

	gProto := a.GetGateway(r)
	if gProto == nil {
		writeErrorResponseXML(ctx, w,
			gerror.GetError(gerror.ErrServerNotInitialized, nil))
		return
	}

	type CreateBucketConfiguration struct {
		LocationConstraint string `xml:"LocationConstraint"`
	}

	var configLocation CreateBucketConfiguration
	if r.Body != nil {
		err := xml.NewDecoder(r.Body).Decode(&configLocation)
		if err != nil {
			if err != io.EOF {
				zlog.ZError().Err(err).Msg("[xml]")
				writeErrorResponseXML(ctx, w,
					gerror.GetError(gerror.ErrMalformedXML, err))
				return
			}
		}
	}

	location := configLocation.LocationConstraint
	amzACL := r.Header.Get("x-amz-acl")
	amzGrantRead := r.Header.Get("x-amz-grant-read")
	amzGrantWrite := r.Header.Get("x-amz-grant-write")
	amzGrantReadAcp := r.Header.Get("x-amz-grant-read-acp")
	amzGrantWriteAcp := r.Header.Get("x-amz-grant-write-acp")
	amzGrantFullControl := r.Header.Get("x-amz-grant-full-control")

	input := &s3.CreateBucketInput{
		ACL:              aws.String(amzACL),
		Bucket:           aws.String(bucket),
		GrantFullControl: aws.String(amzGrantFullControl),
		GrantRead:        aws.String(amzGrantRead),
		GrantReadACP:     aws.String(amzGrantReadAcp),
		GrantWrite:       aws.String(amzGrantWrite),
		GrantWriteACP:    aws.String(amzGrantWriteAcp),
	}

	if location != "" {
		input.CreateBucketConfiguration = &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(location),
		}
	}

	output, resp, err := gProto.CreateBucketWithContext(ctx, input)
	if err != nil {
		writeS3Header(w, resp.Header)
		writeErrorResponseXML(ctx, w, err)
		return
	}

	writeS3Header(w, resp.Header)
	if output != nil && output.Location != nil {
		w.Header().Set("Location", *output.Location)
	}
	return
}

// DeleteBucket https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/RESTBucketDELETE.html
//
// Deletes the bucket named in the URI.
// All objects (including all object versions and delete markers) in the bucket must be deleted before the bucket itself can be deleted.
func (a *API) DeleteBucket(w http.ResponseWriter, r *http.Request) {

	ctx := newContext(w, r, "DeleteBucket")

	vars := mux.Vars(r)
	bucket := vars["bucket"]

	zlog.ZDebug().Str("Bucket", bucket).Str("Method", "DeleteBucket").Msg("[debug]")

	if errCode := a.Auth(r); errCode != gerror.ErrNone {
		if errCode == gerror.ErrRequestTimeTooSkewed {
			writeErrorRequestTimeTooSkewed(ctx, w, r)
			return
		}
		writeErrorResponseXML(ctx, w,
			gerror.GetError(errCode, nil))
		return
	}

	gProto := a.GetGateway(r)
	if gProto == nil {
		writeErrorResponseXML(ctx, w,
			gerror.GetError(gerror.ErrServerNotInitialized, nil))
		return
	}

	_, resp, err := gProto.DeleteBucketWithContext(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucket)})
	if err != nil {
		writeS3Header(w, resp.Header)
		writeErrorResponseXML(ctx, w, err)
		return
	}
	// writeSuccessResponseHeadersOnly(w)
	writeS3Header(w, resp.Header)
	return
}

// ListBuckets https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/RESTServiceGET.html
//
// This implementation of the GET operation returns a list of all buckets owned by the authenticated sender of the request.
// To authenticate a request, you must use a valid AWS Access Key ID that is registered with Amazon S3.
// Anonymous requests cannot list buckets, and you cannot list buckets that you did not create.
func (a *API) ListBuckets(w http.ResponseWriter, r *http.Request) {

	ctx := newContext(w, r, "ListBuckets")

	zlog.ZDebug().Str("Method", "ListBuckets").Msg("[debug]")

	if errCode := a.Auth(r); errCode != gerror.ErrNone {
		if errCode == gerror.ErrRequestTimeTooSkewed {
			writeErrorRequestTimeTooSkewed(ctx, w, r)
			return
		}
		writeErrorResponseXML(ctx, w,
			gerror.GetError(errCode, nil))
		return
	}

	gProto := a.GetGateway(r)
	if gProto == nil {
		writeErrorResponseXML(ctx, w,
			gerror.GetError(gerror.ErrServerNotInitialized, nil))
		return
	}

	output, resp, err := gProto.ListBucketsWithContext(ctx, &s3.ListBucketsInput{})
	if err != nil {
		writeS3Header(w, resp.Header)
		writeErrorResponseXML(ctx, w, err)
		return
	}

	// 因为解析的原因，所以重新编辑
	tmpR := TmpListBuckets{
		Buckets: output.Buckets,
		Owner:   output.Owner,
	}

	writeS3Header(w, resp.Header)
	formatWriteXML(w, http.StatusOK, "ListAllMyBucketsResult", tmpR, true)
	return
}

type TmpListBuckets struct {
	Buckets []*s3.Bucket `xml:"Buckets>Bucket"`
	Owner   *s3.Owner    `xml:"Owner"`
}
