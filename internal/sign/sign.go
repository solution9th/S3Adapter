// Package sign implements AWS Signature Version 4
package sign

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/solution9th/S3Adapter/internal/gerror"

	"github.com/haozibi/zlog"
)

const (
	TimeISO8601BasicFormat      = "20060102T150405Z"
	TimeISO8601BasicFormatShort = "20060102"
	MaxAllowedSkewMilliseconds  = 900000
)

var (
	Debug                   = false
	ServiceName             = "s3"
	ErrMissDateParams       = errors.New("miss date params")
	ErrRequestTimeTooSkewed = errors.New("The difference between the request time and the current time is too large.")
)

type SignV4 struct {
	name           string
	ak, sk, region string
	w              *strings.Builder
	r              *http.Request
	debug          bool
}

// NewSignV4 new signV4
func NewSignV4(ak, sk, region string, r *http.Request) *SignV4 {
	if ak == "" || sk == "" || r == nil {
		panic(errors.New("miss params"))
	}

	builder := &strings.Builder{}
	builder.Grow(2048)

	return &SignV4{
		name:   ServiceName,
		ak:     ak,
		sk:     sk,
		region: region,
		w:      builder,
		r:      r,
		debug:  Debug,
	}
}

func (s *SignV4) creds(t time.Time) string {
	return t.Format(TimeISO8601BasicFormatShort) + "/" + s.region + "/" + s.name + "/aws4_request"
}

// WriteSignature 把签名写入到 Request 中
// func (s *SignV4) WriteSignature() (err error) {

// 	sign, err := s.GenSignature()
// 	if err != nil {
// 		return err
// 	}

// 	s.r.Header.Set("Authorization", sign)
// 	return nil
// }

// GenSignature 生成签名，如果不存在 Date 会自动创建
// func (s *SignV4) GenSignature() (sign string, err error) {

// 	date := s.r.Header.Get("X-Amz-Date")
// 	t := time.Now().UTC()

// 	if date == "" {
// 		date = s.r.Header.Get("Date")
// 	}

// 	if date != "" {
// 		format := http.TimeFormat

// 		if s.r.Header.Get("X-Amz-Date") != "" {
// 			format = TimeISO8601BasicFormat
// 		}
// 		t, err = time.Parse(format, date)
// 		if err != nil {
// 			return "", err
// 		}
// 	}

// 	s.r.Header.Set("Host", s.r.Host)
// 	return s.signature(t, nil)
// }

// Signature 根据请求生成签名，用于验证
func (s *SignV4) Signature(now time.Time) (string, gerror.APIErrorCode) {

	date := s.r.Header.Get("X-Amz-Date")
	if date == "" {
		date = s.r.Header.Get("Date")
	}

	if date == "" {
		return "", gerror.ErrAllAccessDisabled
	}

	var t time.Time

	format := http.TimeFormat

	if s.r.Header.Get("X-Amz-Date") != "" {
		format = TimeISO8601BasicFormat
	}

	t, err := time.Parse(format, date)
	if err != nil {
		zlog.ZError().Str("Method", "time parse").Msg(err.Error())
		return "", gerror.ErrAllAccessDisabled
	}

	return s.signature(now, t, nil)
}

// Verify 验证签名
func (s *SignV4) Verify(nowTime time.Time, authorization string) gerror.APIErrorCode {

	a, err := NewAuthSign(authorization)
	if err != nil {
		zlog.ZDebug().Str("Method", "NewAuthSign").Msg(err.Error())
		return gerror.ErrAllAccessDisabled
	}

	if a.GetAccessKey() != s.ak {
		return gerror.ErrAllAccessDisabled
	}

	date := s.r.Header.Get("X-Amz-Date")
	if date == "" {
		date = s.r.Header.Get("Date")
	}

	if date == "" {
		zlog.ZError().Msg("miss date")
		return gerror.ErrAllAccessDisabled
	}

	var t time.Time

	format := http.TimeFormat

	if s.r.Header.Get("X-Amz-Date") != "" {
		format = TimeISO8601BasicFormat
	}

	t, err = time.ParseInLocation(format, date, time.UTC)
	// t, err = time.Parse(format, date)
	if err != nil {
		zlog.ZDebug().Str("Method", "NewAuthSign").Msg(err.Error())
		return gerror.ErrAllAccessDisabled
	}

	got, errCode := s.signature(nowTime, t, a.GetSignedHeaders())
	if err != nil {
		zlog.ZDebug().Str("Method", "NewAuthSign").Msg(err.Error())
		return errCode
	}

	ok := got == authorization

	if !ok && s.debug {
		zlog.ZDebug().Int("Step", 4).Str("Got", got).Str("Want", authorization).Bool("Result", ok).Msg("[Sign V4]")
	}

	if ok {
		return gerror.ErrNone
	}

	return gerror.ErrAllAccessDisabled
}

func (s *SignV4) signature(nowTime, t time.Time, headers []string) (sign string, e gerror.APIErrorCode) {

	if t.After(nowTime.Add(MaxAllowedSkewMilliseconds*time.Millisecond)) ||
		t.Before(nowTime.Add(-1*MaxAllowedSkewMilliseconds*time.Millisecond)) {
		return "", gerror.ErrRequestTimeTooSkewed
	}

	SigningKey := s.buildSignature(t)

	h := hmac.New(sha256.New, SigningKey)

	// Step 2: StringToSign
	if s.debug {
		buf := new(bytes.Buffer)
		s.buildStringToSign(buf, t, headers, false)
		fmt.Println("> StringToSign")
		fmt.Println(buf.String())
		fmt.Println("> StringToSign over")
		h.Write(buf.Bytes())
	} else {
		s.buildStringToSign(h, t, headers, false)
	}

	s.w.WriteString("AWS4-HMAC-SHA256 ")
	s.w.WriteString("Credential=" + s.ak + "/" + s.creds(t) + ", ")
	s.w.WriteString("SignedHeaders=")

	if len(headers) == 0 {
		s.buildHeaderList(s.w)
	} else {
		s.w.WriteString(strings.Join(headers, ";"))
	}

	s.w.WriteString(", ")
	s.w.WriteString("Signature=" + fmt.Sprintf("%x", h.Sum(nil)))

	return s.w.String(), gerror.ErrNone
}

func (s *SignV4) buildHeaderList(w io.Writer) {
	i, a := 0, make([]string, len(s.r.Header))
	for k := range s.r.Header {
		a[i] = strings.ToLower(k)
		i++
	}
	sort.Strings(a)
	for k, v := range a {
		if k > 0 {
			w.Write([]byte{';'})
		}
		w.Write([]byte(v))
	}
}

func (s *SignV4) buildSignature(t time.Time) []byte {

	var (
		value       = []byte("AWS4" + s.sk)
		YYYYMMDD    = []byte(t.Format(TimeISO8601BasicFormatShort))
		awsRegion   = []byte(s.region)
		awsService  = []byte(s.name)
		aws4Request = []byte("aws4_request")
	)

	DataKey := s.hmacSHA256(value, YYYYMMDD)
	DateRegionKey := s.hmacSHA256(DataKey, awsRegion)
	DateRegionServiceKey := s.hmacSHA256(DateRegionKey, awsService)
	SigningKey := s.hmacSHA256(DateRegionServiceKey, aws4Request)

	if s.debug {
		zlog.ZDebug().Int("Step", 3).Str("SigningKey", string(SigningKey)).Msg("[Sign V4]")
	}

	return SigningKey
}

// Step 2: StringToSign
func (s *SignV4) buildStringToSign(w io.Writer, t time.Time, headers []string, isPresigned bool) {

	w.Write([]byte("AWS4-HMAC-SHA256"))
	w.Write([]byte("\n"))
	w.Write([]byte(t.Format(TimeISO8601BasicFormat)))
	w.Write([]byte("\n"))
	w.Write([]byte(s.creds(t)))
	w.Write([]byte("\n"))

	h := sha256.New()
	if s.debug {
		buf := new(bytes.Buffer)
		s.writeRequest(buf, headers, isPresigned)
		fmt.Println("> Canonical Request")
		fmt.Println(buf.String())
		fmt.Println("> Canonical Request over")
		h.Write(buf.Bytes())
	} else {
		s.writeRequest(h, headers, isPresigned)
	}

	fmt.Fprintf(w, "%x", h.Sum(nil))
}

// CanonicalRequest
func (s *SignV4) writeRequest(w io.Writer, headers []string, isPresigned bool) {

	if s.r.Header.Get("Host") == "" {
		s.r.Header.Add("Host", s.r.Host)
	}

	w.Write([]byte(s.r.Method))
	w.Write([]byte("\n"))
	s.buildURI(w)
	w.Write([]byte("\n"))
	s.buildQuery(w, isPresigned)
	w.Write([]byte("\n"))
	s.buildHeader(w, headers)
	w.Write([]byte("\n"))
	w.Write([]byte("\n"))
	if len(headers) == 0 {
		s.buildHeaderList(w)
	} else {
		w.Write([]byte(strings.Join(headers, ";")))
	}
	w.Write([]byte("\n"))
	if isPresigned {
		w.Write([]byte("UNSIGNED-PAYLOAD"))
	} else {
		s.buildBody(w)
	}
}

func (s *SignV4) buildURI(w io.Writer) {
	path := s.r.URL.RequestURI()
	if s.r.URL.RawQuery != "" {
		path = path[:len(path)-len(s.r.URL.RawQuery)-1]
	}
	slash := strings.HasSuffix(path, "/")
	path = filepath.Clean(path)
	if path != "/" && slash {
		path += "/"
	}
	path = s.uriEncode(path, false)
	w.Write([]byte(path))
}

func (s *SignV4) buildQuery(w io.Writer, isPresigned bool) {
	var a []string
	for k, vs := range s.r.URL.Query() {
		if k == "X-Amz-Signature" && isPresigned {
			continue
		}
		k = url.QueryEscape(k)
		for _, v := range vs {
			if v == "" {
				a = append(a, k+"=")
			} else {
				v = url.QueryEscape(v)
				a = append(a, k+"="+v)
			}
		}
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			w.Write([]byte{'&'})
		}
		w.Write([]byte(s))
	}
}

func (s *SignV4) buildHeader(w io.Writer, headers []string) {
	if len(headers) == 0 {
		for k := range s.r.Header {
			headers = append(headers, k)
		}
	}

	i, a := 0, make([]string, len(headers))
	for _, v := range headers {

		v = http.CanonicalHeaderKey(v)
		value := (map[string][]string)(s.r.Header)[v]
		sort.Strings(value)

		a[i] = strings.ToLower(v) + ":" + strings.Join(value, ",")
		i++
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			w.Write([]byte("\n"))
		}
		io.WriteString(w, s)
	}
}

func (s *SignV4) buildBody(w io.Writer) {
	var b []byte
	// If the payload is empty, use the empty string as the input to the SHA256 function
	// https://docs.aws.amazon.com/zh_cn/general/latest/gr/sigv4-create-canonical-request.html
	if s.r.Body == nil {
		b = []byte("")
	} else {
		var err error
		b, err = ioutil.ReadAll(s.r.Body)
		if err != nil {
			panic(err)
		}
		s.r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	}

	h := sha256.New()
	h.Write(b)
	fmt.Fprintf(w, "%x", h.Sum(nil))
}

func (s *SignV4) hmacSHA256(key, date []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(date)
	return h.Sum(nil)
}

func (s *SignV4) uriEncode(uri string, encodeSlash bool) string {

	input := []byte(uri)
	r := make([]byte, 0)
	for i := 0; i < len(input); i++ {
		ch := input[i]
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') ||
			(ch >= '0' && ch <= '9') || ch == '_' || ch == '-' ||
			ch == '~' || ch == '.' {
			r = append(r, ch)
		} else if ch == '/' {
			if encodeSlash {
				r = append(r, []byte{'%', '2', 'F'}...)
			} else {
				r = append(r, ch)
			}
		} else {
			t := "%" + strings.ToUpper(hex.EncodeToString([]byte{ch}))
			r = append(r, []byte(t)...)
		}
	}
	return string(r)
}
