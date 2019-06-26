package app

import (
	"github.com/solution9th/S3Adapter/internal/gerror"

	"github.com/haozibi/zlog"
)

type CreateApplicationConfiguration struct {
	AccessKey string `xml:"AccessKey"`
	SecretKey string `xml:"SecretKey"`
	Region    string `xml:"Region"`
	Engine    string `xml:"Engine"`
	AppName   string `xml:"AppName"`
	AppRemark string `xml:"AppRemark"`
}

func (a *API) deleteInfo(oak, osk string) error {

	err := a.DB.DeleteInfo(oak, osk)
	if err != nil {
		zlog.ZError().Str("Method", "deleteInfo").Msg(err.Error())
	}
	return err
}

func (a *API) saveInfo(p CreateApplicationConfiguration) (ak, sk string, errCode gerror.APIErrorCode) {

	num, err := a.DB.CountInfo(p.AccessKey, p.SecretKey, p.Engine)
	if err != nil {
		zlog.ZError().Str("Method", "countInfo").Msg(err.Error())
		return "", "", gerror.ErrInternalError
	}

	if num >= 1 {
		return "", "", gerror.ErrAccessKeyCreated
	}

	if p.Engine == "" || p.AccessKey == "" || p.SecretKey == "" || p.Region == "" {
		return "", "", gerror.ErrInvalidRequestParameter
	}

	ak = genAccessKey()
	sk = genSecretKey()

	data := make(map[string]interface{})
	data["os_access_key"] = ak
	data["os_screct_key"] = sk
	data["engine_type"] = p.Engine
	data["engine_access_key"] = p.AccessKey
	data["engine_secret_key"] = p.SecretKey
	data["app_name"] = p.AppName
	data["app_remark"] = p.AppRemark
	data["engine_region"] = p.Region

	_, err = a.DB.SaveInfo(data)

	if err != nil {
		zlog.ZError().Str("Method", "saveInfo").Msg(err.Error())
		return "", "", gerror.ErrInternalError
	}

	return
}
