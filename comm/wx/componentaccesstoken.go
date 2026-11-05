package wx

import (
	"errors"
	"fmt"

	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/httputils"
	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/log"
	wxbase "github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/wx/base"
	"github.com/WeixinCloud/wxcloudrun-wxcomponent/db/model"
)

type componentAccessTokenReq struct {
	ComponentAppid        string `wx:"component_appid"`
	ComponentAppSecret    string `wx:"component_appsecret"`
	ComponentVerifyTicket string `wx:"component_verify_ticket"`
}

type componentAccessTokenResp struct {
	ComponentAccessToken string `wx:"component_access_token"`
	ExpiresIn            uint64 `wx:"expires_in"`
}

// GetComponentAccessToken 获取ComponentAccessToken
func GetComponentAccessToken() (string, error) {
	return getAccessTokenWithRetry(wxbase.GetAppid(), model.WXTOKENTYPE_OWN)
}

func getNewComponentAccessToken() (string, error) {
	ticket := wxbase.GetTicket()
	if len(ticket) == 0 {
		log.Errorf("component token request skipped: appid=%s ticket is empty", wxbase.GetAppid())
		return "", errors.New("empty ticket")
	}
	log.Infof("requesting component access token: appid=%s ticket_length=%d", wxbase.GetAppid(), len(ticket))
	req := componentAccessTokenReq{
		ComponentAppid:        wxbase.GetAppid(),
		ComponentAppSecret:    wxbase.GetSecret(),
		ComponentVerifyTicket: ticket,
	}
	var resp componentAccessTokenResp
	requestBody, err := WxJson.Marshal(req)
	if err != nil {
		log.Errorf("component token request encode failed: appid=%s err=%v", wxbase.GetAppid(), err)
		return "", err
	}
	_, body, err := httputils.RawPost(GetRawWxApiUrl("/cgi-bin/component/api_component_token", ""), requestBody, "application/json")
	if err != nil {
		log.Errorf("component token HTTP request failed: appid=%s err=%v", wxbase.GetAppid(), err)
		return "", err
	}
	var wxErr WxCommError
	if err := WxJson.Unmarshal(body, &wxErr); err != nil {
		log.Errorf("component token response parse failed: appid=%s err=%v", wxbase.GetAppid(), err)
		return "", err
	}
	if err := WxJson.Unmarshal(body, &resp); err != nil {
		log.Errorf("component token response parse failed: appid=%s err=%v", wxbase.GetAppid(), err)
		return "", err
	}
	if wxErr.ErrCode != 0 {
		log.Errorf("WeChat rejected component token request: appid=%s errcode=%d errmsg=%s", wxbase.GetAppid(), wxErr.ErrCode, wxErr.ErrMsg)
		return "", fmt.Errorf("WxErrCode != 0, errcode=%d errmsg=%s", wxErr.ErrCode, wxErr.ErrMsg)
	}
	if resp.ComponentAccessToken == "" {
		return "", errors.New("empty component access token in WeChat response")
	}
	log.Infof("component access token received: appid=%s expires_in=%d", wxbase.GetAppid(), resp.ExpiresIn)
	return resp.ComponentAccessToken, nil
}
