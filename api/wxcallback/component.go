package wxcallback

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/errno"
	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/log"
	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/wx"

	wxbase "github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/wx/base"
	"github.com/WeixinCloud/wxcloudrun-wxcomponent/db/dao"
	"github.com/WeixinCloud/wxcloudrun-wxcomponent/db/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type wxCallbackComponentRecord struct {
	CreateTime int64  `json:"CreateTime"`
	InfoType   string `json:"InfoType"`
}

func componentHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		handleComponentVerification(c)
		return
	}

	// 记录到数据库
	body, _ := ioutil.ReadAll(c.Request.Body)
	body, err := decodeComponentCallback(c, body)
	if err != nil {
		log.Errorf("decode component callback: %v", err)
		c.String(http.StatusBadRequest, "invalid callback")
		return
	}
	var json wxCallbackComponentRecord
	if err := binding.JSON.BindBody(body, &json); err != nil {
		log.Errorf("parse component callback: %v", err)
		c.JSON(http.StatusBadRequest, errno.ErrInvalidParam.WithData(err.Error()))
		return
	}
	r := model.WxCallbackComponentRecord{
		CreateTime:  time.Unix(json.CreateTime, 0),
		ReceiveTime: time.Now(),
		InfoType:    json.InfoType,
		PostBody:    string(body),
	}
	if json.CreateTime == 0 {
		r.CreateTime = time.Unix(1, 0)
	}
	if err := dao.AddComponentCallBackRecord(&r); err != nil {
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}

	// 处理授权相关的消息
	switch json.InfoType {
	case "component_verify_ticket":
		err = ticketHandler(&body)
	case "authorized":
		fallthrough
	case "updateauthorized":
		err = newAuthHander(&body)
	case "unauthorized":
		err = unAuthHander(&body)
	}
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}

	// 转发到用户配置的地址
	var proxyOpen bool
	proxyOpen, err = proxyCallbackMsg(json.InfoType, "", "", string(body), c)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	if !proxyOpen {
		c.String(http.StatusOK, "success")
	}
}

// componentXMLMessage is the decrypted message sent inside WeChat's AES envelope.
type componentXMLMessage struct {
	XMLName                      xml.Name `xml:"xml" json:"-"`
	CreateTime                   int64    `xml:"CreateTime" json:"CreateTime"`
	InfoType                     string   `xml:"InfoType" json:"InfoType"`
	ComponentVerifyTicket        string   `xml:"ComponentVerifyTicket" json:"ComponentVerifyTicket"`
	AuthorizerAppid              string   `xml:"AuthorizerAppid" json:"AuthorizerAppid"`
	AuthorizationCode            string   `xml:"AuthorizationCode" json:"AuthorizationCode"`
	AuthorizationCodeExpiredTime int64    `xml:"AuthorizationCodeExpiredTime" json:"AuthorizationCodeExpiredTime"`
}

type ticketRecord struct {
	ComponentVerifyTicket string `json:"ComponentVerifyTicket"`
}

func ticketHandler(body *[]byte) error {
	var record ticketRecord
	if err := binding.JSON.BindBody(*body, &record); err != nil {
		return err
	}
	log.Infof("[new ticket] length=%d", len(record.ComponentVerifyTicket))
	if err := wxbase.SetTicket(record.ComponentVerifyTicket); err != nil {
		return err
	}
	return nil
}

type newAuthRecord struct {
	CreateTime                   int64  `json:"CreateTime"`
	AuthorizerAppid              string `json:"AuthorizerAppid"`
	AuthorizationCode            string `json:"AuthorizationCode"`
	AuthorizationCodeExpiredTime int64  `json:"AuthorizationCodeExpiredTime"`
}

func newAuthHander(body *[]byte) error {
	var record newAuthRecord
	var err error
	var refreshtoken string
	var appinfo wx.AuthorizerInfoResp
	if err = binding.JSON.BindBody(*body, &record); err != nil {
		return err
	}
	if refreshtoken, err = queryAuth(record.AuthorizationCode); err != nil {
		return err
	}
	if err = wx.GetAuthorizerInfo(record.AuthorizerAppid, &appinfo); err != nil {
		return err
	}
	if err = dao.CreateOrUpdateAuthorizerRecord(&model.Authorizer{
		Appid:         record.AuthorizerAppid,
		AppType:       appinfo.AuthorizerInfo.AppType,
		ServiceType:   appinfo.AuthorizerInfo.ServiceType.Id,
		NickName:      appinfo.AuthorizerInfo.NickName,
		UserName:      appinfo.AuthorizerInfo.UserName,
		HeadImg:       appinfo.AuthorizerInfo.HeadImg,
		QrcodeUrl:     appinfo.AuthorizerInfo.QrcodeUrl,
		PrincipalName: appinfo.AuthorizerInfo.PrincipalName,
		RefreshToken:  refreshtoken,
		FuncInfo:      appinfo.AuthorizationInfo.StrFuncInfo,
		VerifyInfo:    appinfo.AuthorizerInfo.VerifyInfo.Id,
		AuthTime:      time.Unix(record.CreateTime, 0),
	}); err != nil {
		return err
	}
	return nil
}

type queryAuthReq struct {
	ComponentAppid    string `wx:"component_appid"`
	AuthorizationCode string `wx:"authorization_code"`
}

type authorizationInfo struct {
	AuthorizerRefreshToken string `wx:"authorizer_refresh_token"`
}
type queryAuthResp struct {
	AuthorizationInfo authorizationInfo `wx:"authorization_info"`
}

func queryAuth(authCode string) (string, error) {
	req := queryAuthReq{
		ComponentAppid:    wxbase.GetAppid(),
		AuthorizationCode: authCode,
	}
	var resp queryAuthResp
	_, body, err := wx.PostWxJsonWithComponentToken("/cgi-bin/component/api_query_auth", "", req)
	if err != nil {
		return "", err
	}
	if err := wx.WxJson.Unmarshal(body, &resp); err != nil {
		log.Errorf("Unmarshal err, %v", err)
		return "", err
	}
	return resp.AuthorizationInfo.AuthorizerRefreshToken, nil
}

type unAuthRecord struct {
	CreateTime      int64  `json:"CreateTime"`
	AuthorizerAppid string `json:"AuthorizerAppid"`
}

func unAuthHander(body *[]byte) error {
	var record unAuthRecord
	var err error
	if err = binding.JSON.BindBody(*body, &record); err != nil {
		log.Errorf("bind err %v", err)
		return err
	}
	if err := dao.DelAuthorizerRecord(record.AuthorizerAppid); err != nil {
		log.Errorf("DelAuthorizerRecord err %v", err)
		return err
	}
	return nil
}
