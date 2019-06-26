package app

import (
	"net/http"

	"github.com/solution9th/S3Adapter/internal/gerror"
	"github.com/solution9th/S3Adapter/internal/to"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	"github.com/haozibi/zlog"
)

// HeadObject https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/RESTObjectHEAD.html
//
// The HEAD operation retrieves metadata from an object without returning the object itself.
// This operation is useful if you are interested only in an object's metadata.
// To use HEAD, you must have READ access to the object.
// A HEAD request has the same options as a GET operation on an object.
// The response is identical to the GET response except that there is no response body.
func (a *API) HeadObject(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(w, r, "HeadObject")

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	object := vars["object"]

	zlog.ZDebug().Str("Object", bucket).Str("Method", "HeadObject").Msg("[debug]")

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

	input := &s3.HeadObjectInput{}
	errName, err := to.UnmarshalRequest(ctx, r, input)
	if err != nil {
		zlog.ZInfo().Str("Method", "UnmarshalRequest").Str("Field", errName).Msg("[To]" + err.Error())
		// writeErrorResponseXML(ctx, w,
		// 	gerror.GetError(gerror.ErrInvalidRequestParameter, nil))
		// return
	}
	input.Bucket = aws.String(bucket)
	input.Key = aws.String(object)

	err = input.Validate()
	if err != nil {
		writeErrorResponseXML(ctx, w,
			gerror.GetError(gerror.ErrInvalidRequestParameter, nil))
		return
	}

	output, resp, err := gProto.HeadObjectWithContext(ctx, input)
	if err != nil {
		writeS3Header(w, resp.Header)
		writeErrorResponseXML(ctx, w, err)
		return
	}
	writeS3Header(w, resp.Header)
	to.MarshalResponse(ctx, w, output)
}

// PutObject https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/RESTObjectPUT.html
//
// This implementation of the PUT operation adds an object to a bucket.
// You must have WRITE permissions on a bucket to add an object to it.
// Amazon S3 never adds partial objects; if you receive a success response, Amazon S3 added the entire object to the bucket.
// Amazon S3 is a distributed system. If it receives multiple write requests for the same object simultaneously, it overwrites all but the last object written.
// Amazon S3 does not provide object locking; if you need this, make sure to build it into your application layer or use versioning instead.
func (a *API) PutObject(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(w, r, "PutObject")

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	object := vars["object"]

	zlog.ZDebug().Str("Object", bucket).Str("Method", "PutObject").Msg("[debug]")

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

	input := &s3.PutObjectInput{}
	_, err := to.UnmarshalRequest(ctx, r, input)
	if err != nil {
		panic(err)
	}
	input.Bucket = aws.String(bucket)
	input.Key = aws.String(object)

	err = input.Validate()
	if err != nil {
		panic(err)
	}

	output, resp, err := gProto.PutObjectWithContext(ctx, input)
	if err != nil {
		writeS3Header(w, resp.Header)
		writeErrorResponseXML(ctx, w, err)
		return
	}
	writeS3Header(w, resp.Header)
	to.MarshalResponse(ctx, w, output)
	// ReflectStToRep(output, w)
}

// GetObject https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/RESTObjectGET.html
//
// This implementation of the GET operation retrieves objects from Amazon S3.
// To use GET, you must have READ access to the object.
// If you grant READ access to the anonymous user, you can return the object without using an authorization header.
func (a *API) GetObject(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(w, r, "GetObject")

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	object := vars["object"]

	zlog.ZDebug().Str("Object", bucket).Str("Method", "GetObject").Msg("[debug]")

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

	input := &s3.GetObjectInput{}
	errName, err := to.UnmarshalRequest(ctx, r, input)
	if err != nil {
		zlog.ZInfo().Str("Method", "UnmarshalRequest").Str("Field", errName).Msg("[To]" + err.Error())
		// writeErrorResponseXML(ctx, w,
		// 	gerror.GetError(gerror.ErrInvalidRequestParameter, nil))
		// return
	}

	input.Bucket = aws.String(bucket)
	input.Key = aws.String(object)
	err = input.Validate()
	if err != nil {
		writeErrorResponseXML(ctx, w,
			gerror.GetError(gerror.ErrInvalidRequestParameter, nil))
		return
	}

	output, resp, err := gProto.GetObjectWithContext(ctx, input)
	if err != nil {
		writeS3Header(w, resp.Header)
		writeErrorResponseXML(ctx, w, err)
		return
	}
	writeS3Header(w, resp.Header)
	to.MarshalResponse(ctx, w, output)
}

// DeleteObject https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/RESTObjectDELETE.html
//
// The DELETE operation removes the null version (if there is one) of an object and inserts a delete marker, which becomes the current version of the object.
// If there isn't a null version, Amazon S3 does not remove any objects.
func (a *API) DeleteObject(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(w, r, "DeleteObject")

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	object := vars["object"]

	zlog.ZDebug().Str("Object", bucket).Str("Method", "DeleteObject").Msg("[debug]")

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

	input := &s3.DeleteObjectInput{}
	errName, err := to.UnmarshalRequest(ctx, r, input)
	if err != nil {
		zlog.ZInfo().Str("Method", "UnmarshalRequest").Str("Field", errName).Msg("[To]" + err.Error())
	}
	input.Bucket = aws.String(bucket)
	input.Key = aws.String(object)

	err = input.Validate()
	if err != nil {
		writeErrorResponseXML(ctx, w,
			gerror.GetError(gerror.ErrInvalidRequestParameter, nil))
		return
	}

	output, resp, err := gProto.DeleteObjectWithContext(ctx, input)
	if err != nil {
		writeS3Header(w, resp.Header)
		writeErrorResponseXML(ctx, w, err)
		return
	}
	writeS3Header(w, resp.Header)
	_ = output
	to.MarshalResponse(ctx, w, nil)
}

// CopyObject https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/RESTObjectCOPY.html
//
// This implementation of the PUT operation creates a copy of an object that is already stored in Amazon S3.
// A PUT copy operation is the same as performing a GET and then a PUT. Adding the request header,
// x-amz-copy-source, makes the PUT operation copy the source object into the destination bucket.
func (a *API) CopyObject(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(w, r, "CopyObject")

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	object := vars["object"]

	zlog.ZDebug().Str("Object", bucket).Str("Method", "CopyObject").Msg("[debug]")

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

	input := &s3.CopyObjectInput{}
	errName, err := to.UnmarshalRequest(ctx, r, input)
	if err != nil {
		zlog.ZInfo().Str("Method", "UnmarshalRequest").Str("Field", errName).Msg("[To]" + err.Error())
	}
	input.Bucket = aws.String(bucket)
	input.Key = aws.String(object)
	err = input.Validate()
	if err != nil {
		writeErrorResponseXML(ctx, w,
			gerror.GetError(gerror.ErrInvalidRequestParameter, nil))
		return
	}

	output, resp, err := gProto.CopyObjectWithContext(ctx, input)
	if err != nil {
		writeS3Header(w, resp.Header)
		writeErrorResponseXML(ctx, w, err)
		return
	}
	writeS3Header(w, resp.Header)
	to.MarshalResponse(ctx, w, output)
}
