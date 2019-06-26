package app

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/solution9th/S3Adapter/internal/gerror"
	"github.com/solution9th/S3Adapter/internal/sign"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/haozibi/zlog"
)

// mimeType represents various MIME type used API responses.
type mimeType string

const (
	// Means no response type.
	mimeNone mimeType = ""
	// Means response type is JSON.
	mimeJSON mimeType = "application/json"
	// Means response type is XML.
	mimeXML mimeType = "application/xml"
)

func formatWriteXML(w http.ResponseWriter, statusCode int, name string, response interface{}, needXmlns bool) {
	var bytesBuffer bytes.Buffer
	bytesBuffer.WriteString(xml.Header)
	xmlns := ""
	if needXmlns {
		xmlns = ` xmlns="http://s3.amazonaws.com/doc/2006-03-01/"`
	}

	v := reflect.TypeOf(response)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if name == "" {
		name = v.Name()
	}

	e := xml.NewEncoder(&bytesBuffer)
	e.Encode(response)
	body := bytesBuffer.Bytes()
	body = bytes.Replace(body, []byte("<"+v.Name()+">"), []byte("<"+name+xmlns+">"), 1)
	body = bytes.Replace(body, []byte("</"+v.Name()+">"), []byte("</"+name+">"), 1)
	body = bytes.Replace(body, []byte("&#34;"), []byte("&quot;"), -1)
	writeResponse(w, statusCode, body, mimeXML)
}

// XMLResponseError 报错信息
type XMLResponseError struct {
	XMLName                    xml.Name `xml:"Error" json:"-"`
	Code                       string   `xml:"Code"`
	RequestTime                string   `xml:"RequestTime,omitempty"`
	ServerTime                 string   `xml:"ServerTime,omitempty"`
	MaxAllowedSkewMilliseconds int      `xml:"MaxAllowedSkewMilliseconds,omitempty"`
	Message                    string   `xml:"Message"`
	ArgumentName               string   `xml:"ArgumentName,omitempty"`
	ArgumentValue              string   `xml:"ArgumentValue,omitempty"`
	Resource                   string   `xml:"Resource,omitempty"`
	RequestID                  string   `xml:"RequestId,omitempty"`
	HostID                     string   `xml:"HostId,omitempty"`
}

type errArgs func(o *XMLResponseError)

// AddArg XMLResponseError 增加参数某些字段
func AddArg(a string) errArgs {
	return func(o *XMLResponseError) {
		o.ArgumentName = o.ArgumentName + "," + a
		o.ArgumentName = strings.TrimPrefix(o.ArgumentName, ",")
	}
}

func AddArgValue(v string) errArgs {
	return func(o *XMLResponseError) {
		o.ArgumentValue = v
	}
}

func AddResource(r string) errArgs {
	return func(o *XMLResponseError) {
		o.Resource = r
	}
}

func AddRequestTime(r string) errArgs {
	return func(o *XMLResponseError) {
		o.RequestTime = r
	}
}

func AddServerTime(r string) errArgs {
	return func(o *XMLResponseError) {
		o.ServerTime = r
	}
}

func AddMaxAllowedSkewMilliseconds(r int) errArgs {
	return func(o *XMLResponseError) {
		o.MaxAllowedSkewMilliseconds = r
	}
}

func writeErrorRequestTimeTooSkewed(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	date := r.Header.Get("X-Amz-Date")
	if date == "" {
		date = r.Header.Get("Date")
	}

	if date == "" {
		zlog.ZError().Msg(fmt.Sprintf("miss data"))
		writeErrorResponseXML(ctx, w,
			gerror.GetError(gerror.ErrAllAccessDisabled, nil))
		return
	}

	var t time.Time

	format := http.TimeFormat

	if r.Header.Get("X-Amz-Date") != "" {
		format = sign.TimeISO8601BasicFormat
	}

	t, err := time.Parse(format, date)
	if err != nil {
		zlog.ZError().Msg(err.Error())
		writeErrorResponseXML(ctx, w,
			gerror.GetError(gerror.ErrAllAccessDisabled, nil))
		return
	}

	rTime := t.Format(sign.TimeISO8601BasicFormat)
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	writeErrorResponseXML(ctx, w, gerror.GetError(gerror.ErrRequestTimeTooSkewed, nil), AddMaxAllowedSkewMilliseconds(sign.MaxAllowedSkewMilliseconds), AddServerTime(now), AddRequestTime(rTime))
}

func writeErrorResponseXML(ctx context.Context, w http.ResponseWriter, err error, args ...errArgs) {

	requestID := w.Header().Get(responseRequestIDKey)
	HostID := w.Header().Get(responseAMZIDKey)

	re, ok := err.(awserr.RequestFailure)
	if ok {
		xe := &XMLResponseError{
			Code:      re.Code(),
			Message:   re.Message(),
			RequestID: requestID,
			HostID:    HostID,
		}
		for _, a := range args {
			a(xe)
		}
		formatWriteXML(w, re.StatusCode(), "Error", xe, false)
		return
	}

	ee, ok := err.(awserr.Error)
	if ok {
		xe := &XMLResponseError{
			Code:      ee.Code(),
			Message:   ee.Message(),
			RequestID: requestID,
			HostID:    HostID,
		}
		for _, a := range args {
			a(xe)
		}
		formatWriteXML(w, 200, "Error", xe, false)
		return
	}
	formatWriteXML(w, 300, "Error", nil, false)
	return
}

func writeSuccessResponseXML(w http.ResponseWriter, response []byte) {
	writeResponse(w, http.StatusOK, response, mimeXML)
}

func writeErrorResponseHeadersOnly(w http.ResponseWriter, err error) {
	if ee, ok := err.(awserr.RequestFailure); ok {
		writeResponse(w, ee.StatusCode(), nil, mimeNone)
		return
	}

	writeResponse(w, 520, nil, mimeNone)
}

func writeSuccessResponseHeadersOnly(w http.ResponseWriter) {
	writeResponse(w, http.StatusOK, nil, mimeNone)
}

func writeS3Header(w http.ResponseWriter, h http.Header) {
	for k, v := range h {
		for _, j := range v {
			// 替换之前的 Header
			w.Header().Set(k, j)
		}
	}
}

func writeResponse(w http.ResponseWriter, statusCode int, response []byte, mType mimeType) {
	setCommonHeaders(w)
	w.Header().Set("Content-Length", strconv.Itoa(len(response)))
	if mType != mimeNone {
		w.Header().Set("Content-Type", string(mType))
	}
	w.WriteHeader(statusCode)
	if response != nil {
		w.Write(response)
		// w.(http.Flusher).Flush()
	}
}

func setCommonHeaders(w http.ResponseWriter) {

	if w.Header().Get("Date") == "" {
		w.Header().Set("Date", time.Now().Format(http.TimeFormat))
	}

	if w.Header().Get("Server") == "" {
		w.Header().Set("Server", BuildAppName+"/"+BuildVersion)
	}
}
