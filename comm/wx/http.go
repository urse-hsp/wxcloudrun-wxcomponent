package wx

import (
	"fmt"
	"mime/multipart"

	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/config"
	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/httputils"
	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/log"
	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/wx/cloudbasetoken"
	jsoniter "github.com/json-iterator/go"
)

// WxCommError 微信开放平台api通用错误
type WxCommError struct {
	ErrCode int    `wx:"errcode"`
	ErrMsg  string `wx:"errmsg"`
}

var WxJson = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
	TagKey:                 "wx",
}.Froze()

// GetComponentWxApiUrl 拼接微信开放平台的url，带第三方token
func GetComponentWxApiUrl(path string, query string) (string, error) {
	if len(query) > 0 {
		query = "&" + query
	}
	var protocol string
	if config.WxApiConf.UseHttps {
		protocol = "https"
	} else {
		protocol = "http"
	}
	url := fmt.Sprintf("%s://api.weixin.qq.com%s", protocol, path)

	if config.WxApiConf.UseCloudBaseAccessToken {
		token := cloudbasetoken.GetCloudBaseAccessToken()
		if len(token) == 0 {
			log.Errorf("build component api url failed: path=%s cloudbase_access_token is empty, "+
				"UseCloudBaseAccessToken=true but token not loaded (only works inside WeChat CloudBase with open-api service enabled)", path)
		}
		log.Infof("build component api url: path=%s token_mode=cloudbase token_length=%d", path, len(token))
		return fmt.Sprintf("%s?cloudbase_access_token=%s%s",
			url, token, query), nil
	}
	if config.WxApiConf.UseComponentAccessToken {
		token, err := GetComponentAccessToken()
		if err != nil {
			log.Errorf("build component api url failed: path=%s token_mode=component get component_access_token err=%v", path, err)
			return "", err
		}
		if len(token) == 0 {
			log.Errorf("build component api url failed: path=%s token_mode=component component_access_token is empty", path)
		}
		log.Infof("build component api url: path=%s token_mode=component token_length=%d", path, len(token))
		return fmt.Sprintf("%s?component_access_token=%s%s",
			url, token, query), nil
	}
	log.Errorf("build component api url failed: path=%s NO access token mode enabled "+
		"(both UseCloudBaseAccessToken and UseComponentAccessToken are false in comm/config/server.conf), "+
		"request will be sent WITHOUT token and WeChat will return 41001 access_token missing", path)
	return fmt.Sprintf("%s?%s", url, query), nil
}

// GetAuthorizerWxApiUrl 拼接微信开放平台的url，带小程序token
func GetAuthorizerWxApiUrl(appid string, path string, query string) (string, error) {
	if len(query) > 0 {
		query = "&" + query
	}
	token, err := GetAuthorizerAccessToken(appid)
	if err != nil {
		log.Error(err)
		return "", err
	}
	return fmt.Sprintf("https://api.weixin.qq.com%s?access_token=%s%s",
		path, token, query), nil
}

// GetRawWxApiUrl 拼接微信开放平台的url，不带微信令牌
func GetRawWxApiUrl(path string, query string) string {
	if len(query) > 0 {
		query = "?" + query
	}
	return fmt.Sprintf("https://api.weixin.qq.com%s%s", path, query)
}

// postWxJson 向微信开放平台发起post请求 解析结构体中的wx标签
func postWxJson(url string, data interface{}) (*WxCommError, []byte, error) {
	var wxError WxCommError
	var body []byte
	var err error
	jsonByte, _ := WxJson.Marshal(data)
	if body, err = httputils.Post(url, jsonByte, "application/json"); err != nil {
		return &wxError, body, err
	}
	if err = WxJson.Unmarshal(body, &wxError); err != nil {
		log.Errorf("Unmarshal err, %v", err)
		return &wxError, body, err
	}
	if wxError.ErrCode != 0 {
		return &wxError, body, fmt.Errorf("WxErrCode != 0, resp: %v", wxError)
	}
	return &wxError, body, nil
}

// postWxFormData 向微信开放平台发起post请求 解析结构体中的wx标签
func postWxFormData(url string, formFile multipart.File,
	fileName string, fieldName string) (*WxCommError, []byte, error) {
	var wxError WxCommError
	body, err := httputils.PostFormData(url, formFile, fileName, fieldName)
	if err != nil {
		return nil, nil, err
	}
	if err = WxJson.Unmarshal(body, &wxError); err != nil {
		log.Errorf("Unmarshal err, %v", err)
		return &wxError, body, err
	}
	if wxError.ErrCode != 0 {
		return &wxError, body, fmt.Errorf("WxErrCode != 0, resp: %v", wxError)
	}
	return &wxError, body, nil
}

// getWxApi 向微信开放平台发起get请求
func getWxApi(url string) (*WxCommError, []byte, error) {
	var wxError WxCommError
	var body []byte
	var err error
	if body, err = httputils.Get(url); err != nil {
		return &wxError, body, err
	}
	if err = WxJson.Unmarshal(body, &wxError); err != nil {
		log.Errorf("Unmarshal err, %v", err)
		return &wxError, body, err
	}
	if wxError.ErrCode != 0 {
		return &wxError, body, fmt.Errorf("WxErrCode != 0, resp: %v", wxError)
	}
	return &wxError, body, nil
}

// PostWxJsonWithComponentToken 以第三方身份向微信开放平台发起post请求
func PostWxJsonWithComponentToken(path string, query string, data interface{}) (*WxCommError, []byte, error) {
	url, err := GetComponentWxApiUrl(path, query)
	if err != nil {
		return nil, []byte{}, err
	}
	return postWxJson(url, data)
}

// PostWxJsonWithAuthToken 以小程序身份向微信开放平台发起post请求
func PostWxJsonWithAuthToken(appid string, path string, query string, data interface{}) (*WxCommError, []byte, error) {
	url, err := GetAuthorizerWxApiUrl(appid, path, query)
	if err != nil {
		return nil, []byte{}, err
	}
	return postWxJson(url, data)
}

// PostWxJsonWithoutToken 向微信开放平台发起post请求
func PostWxJsonWithoutToken(path string, query string, data interface{}) (*WxCommError, []byte, error) {
	return postWxJson(GetRawWxApiUrl(path, query), data)
}

// GetWxApiWithComponentToken 以第三方身份向微信开放平台发起get请求
func GetWxApiWithComponentToken(path string, query string) (*WxCommError, []byte, error) {
	url, err := GetComponentWxApiUrl(path, query)
	if err != nil {
		return nil, []byte{}, err
	}
	return getWxApi(url)
}

// GetWxApiWithAuthToken 以小程序身份向微信开放平台发起get请求
func GetWxApiWithAuthToken(appid string, path string, query string) (*WxCommError, []byte, error) {
	url, err := GetAuthorizerWxApiUrl(appid, path, query)
	if err != nil {
		return nil, []byte{}, err
	}
	return getWxApi(url)
}

// GetWxApiWithoutToken 向微信开放平台发起get请求
func GetWxApiWithoutToken(path string, query string) (*WxCommError, []byte, error) {
	return getWxApi(GetRawWxApiUrl(path, query))
}

// PostWxFormDataWithAuthToken 以小程序身份向微信开放平台发起post请求
func PostWxFormDataWithAuthToken(appid string, path string, query string,
	formFile multipart.File, fileName string, fieldName string) (*WxCommError, []byte, error) {
	url, err := GetAuthorizerWxApiUrl(appid, path, query)
	if err != nil {
		return nil, []byte{}, err
	}
	return postWxFormData(url, formFile, fileName, fieldName)
}
