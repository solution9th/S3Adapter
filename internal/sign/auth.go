package sign

import (
	"errors"
	"strings"
)

// 示例
// Authorization: AWS4-HMAC-SHA256
// Credential=AKIAIOSFODNN7EXAMPLE/20161208/US/s3/aws4_request,
// SignedHeaders=host;range;x-amz-date,
// Signature=fe5f80f77d5fa3beca038a248ff027d0445342fe2855ddc963176630326f1024

// AuthSigner analyze authorization header and get details
type AuthSigner interface {
	GetAccessKey() string
	GetDate() string
	GetRegion() string
	GetSignedHeaders() []string
	GetSignature() string
	GetServiceName() string
}

const (
	// AWS4HMACSHA256 该字符串指定AWS签名版本4（AWS4）和签名算法（HMAC-SHA256）
	AWS4HMACSHA256 = "AWS4-HMAC-SHA256"

	AWSV2Algorithm = "AWS"

	AWSStreamingContentSHA256 = "STREAMING-AWS4-HMAC-SHA256-PAYLOAD"
)

type authSign struct {
	ak string
	m  map[string]string
}

// NewAuthSign new AuthSigner
func NewAuthSign(s string) (AuthSigner, error) {

	if !strings.HasPrefix(s, AWS4HMACSHA256+" ") {
		return nil, errors.New("got creden error")
	}

	s = strings.TrimSpace(strings.TrimPrefix(s, AWS4HMACSHA256))

	a := &authSign{
		m: make(map[string]string),
	}

	s = strings.Replace(s, " ", "", -1)
	sl := strings.Split(s, ",")
	if len(sl) != 3 {
		return nil, errors.New("got creden error")
	}

	for _, v := range sl {
		sm := strings.Split(v, "=")
		if len(sm) != 2 {
			return nil, errors.New("got creden error")
		}

		a.m[sm[0]] = sm[1]
	}

	return a, nil
}

func (a *authSign) GetSignedHeaders() []string {
	s := a.m["SignedHeaders"]
	if s == "" {
		return nil
	}

	return strings.Split(s, ";")
}

func (a *authSign) GetAccessKey() string {
	return a.getCredential(0)
}

func (a *authSign) GetDate() string {
	return a.getCredential(1)
}

func (a *authSign) GetRegion() string {
	return a.getCredential(2)
}

func (a *authSign) GetServiceName() string {
	return a.getCredential(3)
}

func (a *authSign) getCredential(i int) string {
	s := a.m["Credential"]
	if s == "" || i < 0 || i > 5 {
		return s
	}

	sl := strings.Split(s, "/")
	if len(sl) != 5 {
		return ""
	}

	return sl[i]
}

func (a *authSign) GetSignature() string { return a.m["Signature"] }
