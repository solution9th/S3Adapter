package gerror

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

const (
	// ErrUnsupported 方法不支持
	ErrUnsupported = "ErrUnsupported"

	// ErrMissParams 缺少必要参数
	ErrMissParams = "ErrMissParams"

	// ErrNotAwsErr 不是标准的 aws 错误
	ErrNotAwsErr = "ErrNotAwsErr"
)

var (
	// AWSErrUnsupported 方法不支持
	AWSErrUnsupported = awserr.New(ErrUnsupported, "The Function Not Implemented", errors.New("[core] the function not implemented"))

	// AWSErrMissParams 缺少必要参数
	AWSErrMissParams = awserr.New(ErrMissParams, "Miss Params", errors.New("[core] miss params"))
)

// ===========
// AWS 指定错误
// ===========

var AWSStdErrList = make([]string, 0, 2000)

func init() {
	AWSStdErrList = []string{"AccessDenied", "AccountProblem", "AllAccessDisabled", "AmbiguousGrantByEmailAddress",
		"AuthorizationHeaderMalformed", "BadDigest", "BucketAlreadyExists", "BucketAlreadyOwnedByYou", "BucketNotEmpty",
		"CredentialsNotSupported", "CrossLocationLoggingProhibited", "EntityTooSmall", "EntityTooLarge", "ExpiredToken",
		"IllegalVersioningConfigurationException", "IncompleteBody", "IncorrectNumberOfFilesInPostRequest", "InlineDataTooLarge",
		"InternalError", "InvalidAccessKeyId", "InvalidAddressingHeader", "InvalidArgument", "InvalidBucketName",
		"InvalidBucketState", "InvalidDigest", "InvalidEncryptionAlgorithmError", "InvalidLocationConstraint",
		"InvalidObjectState", "InvalidPart", "InvalidPartOrder", "InvalidPayer", "InvalidPolicyDocument", "InvalidRange",
		"InvalidRequest", "InvalidSecurity", "InvalidSOAPRequest", "InvalidStorageClass", "InvalidTargetBucketForLogging",
		"InvalidToken", "InvalidURI", "KeyTooLongError", "MalformedACLError", "MalformedPOSTRequest", "MalformedXML",
		"MaxMessageLengthExceeded", "MaxPostPreDataLengthExceededError", "MetadataTooLarge", "MethodNotAllowed",
		"MissingAttachment", "MissingContentLength", "MissingRequestBodyError", "MissingSecurityElement", "MissingSecurityHeader",
		"NoLoggingStatusForKey", "NoSuchBucket", "NoSuchBucketPolicy", "NoSuchKey", "NoSuchLifecycleConfiguration", "NoSuchUpload",
		"NoSuchVersion", "NotImplemented", "NotSignedUp", "OperationAborted", "PermanentRedirect", "PreconditionFailed",
		"Redirect", "RestoreAlreadyInProgress", "RequestIsNotMultiPartContent", "RequestTimeout", "RequestTimeTooSkewed",
		"RequestTorrentOfBucketError", "SignatureDoesNotMatch", "ServiceUnavailable", "SlowDown", "TemporaryRedirect",
		"TokenRefreshRequired", "TooManyBuckets", "UnexpectedContent", "UnresolvableGrantByEmailAddress", "UserKeyMustBeSpecified",
	}
}

// GetAWSStdError aws 标准错误 https://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html
func GetAWSStdError(s string, err error) awserr.Error {

	for _, v := range AWSStdErrList {
		if v == s {
			return awserr.New(v, "aws std error: "+v, err)
		}
	}
	return nil
}
