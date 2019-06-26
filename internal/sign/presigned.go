package sign

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/url"
	"time"

	"github.com/solution9th/S3Adapter/internal/gerror"
)

const (
	iso8601Format = "20060102T150405Z"

	// The maximum allowed time difference between the incoming request
	// date and server date during signature verification.
	globalMaxSkewTime = 15 * time.Minute // 15 minutes skew allowed.
)

// 预签名
// https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/API/sigv4-query-string-auth.html

// VerifyURL 验证预签名
func (s *SignV4) VerifyURL(t time.Time) gerror.APIErrorCode {

	p, errCode := s.getURLParamsFromRequest()
	if errCode != gerror.ErrNone {
		return errCode
	}

	utc := t.UTC()

	if p.Date.After(utc.Add(globalMaxSkewTime)) {
		return gerror.ErrRequestNotReadyYet
	}

	if utc.Sub(p.Date) > p.Expires {
		return gerror.ErrExpiredPresignRequest
	}

	newSign, err := s.signatureURL(t, p.SignedHeaders)
	if err != nil {
		return gerror.ErrExpiredPresignRequest
	}

	if newSign == p.Signature {
		return gerror.ErrNone
	}

	return gerror.ErrExpiredPresignRequest
}

type urlParams struct {
	Algorithm     string        `location:"querystring" locationName:"X-Amz-Algorithm"`
	Credential    CredInfo      `location:"querystring" locationName:"X-Amz-Credential"`
	Date          time.Time     `location:"querystring" locationName:"X-Amz-Date" timestampFormat:"iso8601"`
	Expires       time.Duration `location:"querystring" locationName:"X-Amz-Expires"`
	SignedHeaders []string      `location:"querystring" locationName:"X-Amz-SignedHeaders"`
	Signature     string        `location:"querystring" locationName:"X-Amz-Signature"`
}

type CredInfo struct {
	AccessKeyID string
	Date        string
	Region      string
	ServiceName string
}

func (s *SignV4) signatureURL(t time.Time, headers []string) (string, error) {

	SigningKey := s.buildSignature(t)

	h := hmac.New(sha256.New, SigningKey)

	if s.debug {
		buf := new(bytes.Buffer)
		s.buildStringToSign(buf, t, headers, true)
		fmt.Println("> StringToSign")
		fmt.Println(buf.String())
		fmt.Println("> StringToSign over")
		h.Write(buf.Bytes())
	} else {
		s.buildStringToSign(h, t, headers, true)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (s *SignV4) getURLParamsFromRequest() (*urlParams, gerror.APIErrorCode) {

	p := &urlParams{}

	u := s.r.URL.Query()
	errCode := verifyParamsExist(u)
	if errCode != gerror.ErrNone {
		return nil, errCode
	}

	p.Algorithm = u.Get("X-Amz-Algorithm")
	if p.Algorithm != AWS4HMACSHA256 {
		return nil, gerror.ErrInvalidQuerySignatureAlgo
	}

	auth, err := NewAuthSign(AWS4HMACSHA256 + " Credential=" + u.Get("X-Amz-Credential") + ",SignedHeaders=" + u.Get("X-Amz-SignedHeaders") + ",Signature=" + u.Get("X-Amz-Signature"))
	if err != nil {
		return nil, gerror.ErrInvalidQuerySignatureAlgo
	}

	p.Credential.AccessKeyID = auth.GetAccessKey()
	p.Credential.Date = auth.GetDate()
	p.Credential.Region = auth.GetRegion()
	p.Credential.ServiceName = auth.GetServiceName()
	p.SignedHeaders = auth.GetSignedHeaders()
	p.Signature = auth.GetSignature()

	p.Date, err = time.Parse(iso8601Format, u.Get("X-Amz-Date"))
	if err != nil {
		return nil, gerror.ErrMalformedPresignedDate
	}

	p.Expires, err = time.ParseDuration(u.Get("X-Amz-Expires") + "s")
	if err != nil {
		return nil, gerror.ErrMalformedExpires
	}

	if p.Expires < 0 {
		return nil, gerror.ErrNegativeExpires
	}

	if p.Expires.Seconds() > 604800 {
		return nil, gerror.ErrMaximumExpires
	}

	return p, gerror.ErrNone
}

func verifyParamsExist(query url.Values) gerror.APIErrorCode {
	params := []string{"X-Amz-Algorithm", "X-Amz-Credential", "X-Amz-Signature", "X-Amz-Date", "X-Amz-SignedHeaders", "X-Amz-Expires"}

	for _, v := range params {
		if _, ok := query[v]; !ok {
			return gerror.ErrInvalidQueryParams
		}
	}
	return gerror.ErrNone
}
