package sign

import (
	"net/http"
	"strings"
)

// Authorization type.
type AuthType int

// List of all supported auth types.
const (
	AuthTypeUnknown AuthType = iota
	AuthTypeAnonymous
	AuthTypePresigned
	AuthTypePresignedV2
	AuthTypePostPolicy
	AuthTypeStreamingSigned
	AuthTypeSigned
	AuthTypeSignedV2
	AuthTypeJWT
	AuthTypeSTS
)

// GetRequestAuthType Get request authentication type.
func GetRequestAuthType(r *http.Request) AuthType {
	if isRequestSignatureV2(r) {
		return AuthTypeSignedV2
	} else if isRequestPresignedSignatureV2(r) {
		return AuthTypePresignedV2
	} else if isRequestSignStreamingV4(r) {
		return AuthTypeStreamingSigned
	} else if isRequestSignatureV4(r) {
		return AuthTypeSigned
	} else if isRequestPresignedSignatureV4(r) {
		return AuthTypePresigned
	} else if isRequestPostPolicySignatureV4(r) {
		return AuthTypePostPolicy
	} else if _, ok := r.URL.Query()["Action"]; ok {
		return AuthTypeSTS
	} else if _, ok := r.Header["Authorization"]; !ok {
		return AuthTypeAnonymous
	}

	return AuthTypeUnknown
}

// Verify if request has AWS Signature Version '4'.
func isRequestSignatureV4(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get("Authorization"), AWS4HMACSHA256)
}

// Verify if request has AWS Signature Version '2'.
func isRequestSignatureV2(r *http.Request) bool {
	return (!strings.HasPrefix(r.Header.Get("Authorization"), AWS4HMACSHA256) &&
		strings.HasPrefix(r.Header.Get("Authorization"), AWSV2Algorithm))
}

// Verify if request has AWS PreSign Version '4'.
func isRequestPresignedSignatureV4(r *http.Request) bool {
	_, ok := r.URL.Query()["X-Amz-Credential"]
	return ok
}

// Verify request has AWS PreSign Version '2'.
func isRequestPresignedSignatureV2(r *http.Request) bool {
	_, ok := r.URL.Query()["AWSAccessKeyId"]
	return ok
}

// Verify if request has AWS Post policy Signature Version '4'.
func isRequestPostPolicySignatureV4(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") &&
		r.Method == http.MethodPost
}

// Verify if the request has AWS Streaming Signature Version '4'. This is only valid for 'PUT' operation.
func isRequestSignStreamingV4(r *http.Request) bool {
	return r.Header.Get("x-amz-content-sha256") == AWSStreamingContentSHA256 &&
		r.Method == http.MethodPut
}
