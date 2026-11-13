package admin

import (
	"net/http"

	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/errno"
	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/wx"
	"github.com/gin-gonic/gin"
)

type getUserPrivacyReq struct {
	Appid      string `form:"appid" binding:"required"` // 授权小程序appid
	PrivacyVer int    `form:"privacy_ver,default=2"`    // 隐私保护指引版本号，默认2
}

func getUserPrivacyHandler(c *gin.Context) {
	var req getUserPrivacyReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusOK, errno.ErrInvalidParam.WithData(err.Error()))
		return
	}
	data, err := wx.GetPrivacySetting(req.Appid, req.PrivacyVer)
	if err != nil {
		c.JSON(http.StatusOK, errno.ErrRequestErr.WithData(err.Error()))
		return
	}
	c.JSON(http.StatusOK, errno.OK.WithData(data))
}

type updateUserPrivacyReq struct {
	Appid        string                  `json:"appid" binding:"required"` // 授权小程序appid
	PrivacyVer   int                     `json:"privacy_ver"`              // 隐私保护指引版本号
	OwnerSetting wx.PrivacyOwnerSetting  `json:"owner_setting"`            // 用户隐私设置
	SettingList  []wx.PrivacySettingItem `json:"setting_list"`             // 隐私设置列表
}

func updateUserPrivacyHandler(c *gin.Context) {
	var req updateUserPrivacyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, errno.ErrInvalidParam.WithData(err.Error()))
		return
	}
	if err := wx.SetPrivacySetting(req.Appid, wx.PrivacySettingRequest{
		PrivacyVer:   req.PrivacyVer,
		OwnerSetting: req.OwnerSetting,
		SettingList:  req.SettingList,
	}); err != nil {
		c.JSON(http.StatusOK, errno.ErrRequestErr.WithData(err.Error()))
		return
	}
	c.JSON(http.StatusOK, errno.OK)
}
