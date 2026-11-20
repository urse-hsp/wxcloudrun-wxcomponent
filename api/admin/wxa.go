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

// 获取用户隐私设置
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

// 更新用户隐私设置
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

// 获取发布后生效服务器域名列表
// POST https://api.weixin.qq.com/wxa/get_effective_domain?access_token=ACCESS_TOKEN
func getEffectiveDomain(c *gin.Context) {
	var req getUserPrivacyReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusOK, errno.ErrInvalidParam.WithData(err.Error()))
		return
	}
	data, err := wx.GetEffectiveDomain(req.Appid, req.PrivacyVer)
	if err != nil {
		c.JSON(http.StatusOK, errno.ErrRequestErr.WithData(err.Error()))
		return
	}
	c.JSON(http.StatusOK, errno.OK.WithData(data))
}

type upModifyDomainDirectlyReq struct {
	Appid                string                      `json:"appid" binding:"required"` // 授权小程序appid
	PrivacyVer           int                         `json:"privacy_ver"`              // 隐私保护指引版本号
	ModifyDomainDirectly wx.ModifyDomainDirectlyItem `json:"modify_domain_directly"`
}

// 快速配置小程序服务器域名
// POST https://api.weixin.qq.com/wxa/modify_domain_directly?access_token=ACCESS_TOKEN
func updateModifyDomainDirectly(c *gin.Context) {
	var req upModifyDomainDirectlyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, errno.ErrInvalidParam.WithData(err.Error()))
		return
	}

	if err := wx.UpdateModifyDomainDirectly(req.Appid, req.ModifyDomainDirectly); err != nil {
		c.JSON(http.StatusOK, errno.ErrRequestErr.WithData(err.Error()))
		return
	}
	c.JSON(http.StatusOK, errno.OK)
}
