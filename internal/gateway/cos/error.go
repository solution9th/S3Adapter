package cos

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/haozibi/zlog"
	"github.com/tencentyun/cos-go-sdk-v5"
)

const (
	responseRequestIDKey = "x-amz-request-id"
	responseAMZIDKey     = "x-amz-id-2"
)

// 错误处理，把 cos 错误转变成 s3
func toS3Err(err error) awserr.RequestFailure {
	if err == nil {
		return nil
	}

	e, ok := err.(*cos.ErrorResponse)
	if !ok {
		zlog.ZError().Msg(err.Error())

		return awserr.NewRequestFailure(awserr.New("ErrNotFoundError", "Not Found S3 Error,please see OrgErr", err), http.StatusBadRequest, "")
		// return nil
	}

	RequestID := e.RequestID
	if RequestID == "" {
		RequestID = e.Response.Header.Get("X-Cos-Request-Id")
	}

	stausCode := e.Response.StatusCode

	return awserr.NewRequestFailure(awserr.New(e.Code, e.Message, err), stausCode, RequestID)
}

func getErrResponse(err error) *http.Response {

	e, ok := err.(*cos.ErrorResponse)
	if !ok {
		return nil
	}
	RequestID := e.RequestID
	if RequestID == "" {
		RequestID = e.Response.Header.Get("X-Cos-Request-Id")
	}

	TraceID := e.TraceID
	if TraceID == "" {
		TraceID = e.Response.Header.Get("X-Cos-Trace-Id")
	}

	e.Response.Header.Set(responseAMZIDKey, TraceID)
	e.Response.Header.Set(responseRequestIDKey, RequestID)

	return e.Response
}
