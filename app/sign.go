package app

import (
	"context"
	"net/http"
	"time"

	"github.com/solution9th/S3Adapter/internal"
	"github.com/solution9th/S3Adapter/internal/auth"
	"github.com/solution9th/S3Adapter/internal/db/mysql"
	"github.com/solution9th/S3Adapter/internal/gateway"
	"github.com/solution9th/S3Adapter/internal/gerror"
	"github.com/solution9th/S3Adapter/internal/sign"

	"github.com/haozibi/zlog"
)

const (
	// AWS4HMACSHA256 该字符串指定AWS签名版本4（AWS4）和签名算法（HMAC-SHA256）
	AWS4HMACSHA256 = "AWS4-HMAC-SHA256"
)

type authInfo struct {
	oak, osk, ak, sk, region, engine string
}

func (a *API) getAuthorizationInfo(r *http.Request) *authInfo {
	authStr := r.Header.Get("Authorization")
	if authStr == "" {
		return nil
	}

	zlog.ZDebug().Str("Authorization", authStr).Msg("[Sign]")
	auth, err := sign.NewAuthSign(authStr)
	if err != nil {
		zlog.ZError().Msg(err.Error())
		return nil
	}

	oak := auth.GetAccessKey()
	// region := auth.GetRegion()
	osk, ak, sk, en, region := a.getSecretKeyEngine(oak)
	if sk == "" || en == "" {
		zlog.ZDebug().Str("AK", ak).Str("Region", "region").Msg("[Sign] miss sk")
		return nil
	}
	return &authInfo{
		oak:    oak,
		osk:    osk,
		ak:     ak,
		sk:     sk,
		region: region,
		engine: en,
	}
}

// Auth 验证签名 Authorization，暂时不支持 URL 预签名
//
// Authorization: AWS4-HMAC-SHA256
// Credential=AKIAIOSFODNN7EXAMPLE/20161208/US/s3/aws4_request,
// SignedHeaders=host;range;x-amz-date,
// Signature=fe5f80f77d5fa3beca038a248ff027d0445342fe2855ddc963176630326f1024
func (a *API) Auth(r *http.Request) gerror.APIErrorCode {

	switch sign.GetRequestAuthType(r) {
	case sign.AuthTypeSigned:
		authStr := r.Header.Get("Authorization")
		authInfo, err := sign.NewAuthSign(authStr)
		if err != nil {
			zlog.ZError().Msg(err.Error())
			return gerror.ErrAllAccessDisabled
		}

		if authInfo.GetRegion() != GlobalRegion {
			return gerror.ErrInvalidRegion
		}

		info := a.getAuthorizationInfo(r)
		if info == nil {
			return gerror.ErrAllAccessDisabled
		}

		// 验证接收的请求，所以需要 oak 和 osk，并不是ak和sk
		sv4 := sign.NewSignV4(info.oak, info.osk, GlobalRegion, r)
		return sv4.Verify(time.Now(), authStr)
	case sign.AuthTypePresigned:
		auth, err := sign.NewAuthSign(AWS4HMACSHA256 + " Credential=" + r.URL.Query().Get("X-Amz-Credential") + ",SignedHeaders=" + r.URL.Query().Get("X-Amz-SignedHeaders") + ",Signature=" + r.URL.Query().Get("X-Amz-Signature"))
		if err != nil {
			zlog.ZError().Msg(err.Error())
			return gerror.ErrAllAccessDisabled
		}

		if auth.GetRegion() != GlobalRegion {
			return gerror.ErrInvalidRegion
		}

		oak := auth.GetAccessKey()

		osk, ak, sk, en, _ := a.getSecretKeyEngine(oak)
		if sk == "" || en == "" {
			zlog.ZDebug().Str("AK", ak).Str("Region", "region").Msg("[Sign] miss sk")
			return gerror.ErrAllAccessDisabled
		}

		sv4 := sign.NewSignV4(oak, osk, GlobalRegion, r)
		errCode := sv4.VerifyURL(time.Now())
		return errCode
	}
	return gerror.ErrAllAccessDisabled
}

// GetGateway 获得网关
func (a *API) GetGateway(r *http.Request) gateway.S3Protocol {

	info := a.getAuthorizationInfo(r)
	if info == nil {
		return nil
	}

	g, err := internal.NewGateway(info.engine, auth.Credentials{
		AccessKey: info.ak, SecretKey: info.sk},
		info.region)
	if err != nil {
		zlog.ZError().Msg(err.Error())
		return nil
	}

	return g
}

func writeSignError(w http.ResponseWriter) {
	ctx := context.Background()
	writeErrorResponseXML(ctx, w,
		gerror.GetError(gerror.ErrServerNotInitialized, nil))
}

// 根据 ak 查找 sk
// oak,osk 为此项目的 key
// ak,sk 为 s3 的key
func (a *API) getSecretKeyEngine(oak string) (osk, ak, sk, engine, region string) {

	if oak == "" {
		return "", "", "", "", ""
	}

	mm, err := a.DB.GetInfo(oak)
	if err != nil {
		zlog.ZError().Str("OAK", oak).Msg("[DB] error: " + err.Error())
		return "", "", "", "", ""
	}

	m := mm.(mysql.Info)

	zlog.ZDebug().Str("Engine", m.EngineType).Msg("[Sign]")

	return m.OsScrectKey, m.EngineAccessKey, m.EngineSecretKey, m.EngineType, m.EngineRegion
}
