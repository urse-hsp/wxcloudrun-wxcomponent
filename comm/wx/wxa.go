package wx

import "encoding/json"

type privacySettingGetReq struct {
	PrivacyVer int `wx:"privacy_ver"`
}

// PrivacyOwnerSetting 开发者的联系方式和数据留存设置。
type PrivacyOwnerSetting struct {
	ContactEmail         string `json:"contact_email" wx:"contact_email"`
	ContactPhone         string `json:"contact_phone" wx:"contact_phone"`
	ContactQQ            string `json:"contact_qq" wx:"contact_qq"`
	ContactWeixin        string `json:"contact_weixin" wx:"contact_weixin"`
	ExtFileMediaID       string `json:"ext_file_media_id" wx:"ext_file_media_id"`
	NoticeMethod         string `json:"notice_method" wx:"notice_method"`
	StoreExpireTimestamp string `json:"store_expire_timestamp" wx:"store_expire_timestamp"`
}

// PrivacySettingItem 隐私权限及其说明。
type PrivacySettingItem struct {
	PrivacyKey  string `json:"privacy_key" wx:"privacy_key"`
	PrivacyText string `json:"privacy_text" wx:"privacy_text"`
}

// PrivacySettingRequest 设置小程序用户隐私保护指引的请求参数。
type PrivacySettingRequest struct {
	PrivacyVer   int                  `json:"privacy_ver" wx:"privacy_ver"`
	OwnerSetting PrivacyOwnerSetting  `json:"owner_setting" wx:"owner_setting"`
	SettingList  []PrivacySettingItem `json:"setting_list" wx:"setting_list"`
}

// GetPrivacySetting 获取授权小程序的用户隐私保护指引。
func GetPrivacySetting(appid string, privacyVer int) (json.RawMessage, error) {
	_, body, err := PostWxJsonWithAuthToken(
		appid,
		"/cgi-bin/component/getprivacysetting",
		"",
		privacySettingGetReq{PrivacyVer: privacyVer},
	)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(body), nil
}

// SetPrivacySetting 更新授权小程序的用户隐私保护指引。
func SetPrivacySetting(appid string, req PrivacySettingRequest) error {
	_, _, err := PostWxJsonWithAuthToken(appid, "/cgi-bin/component/setprivacysetting", "", req)
	return err
}

// GetEffectiveDomain 获取发布后生效服务器域名
func GetEffectiveDomain(appid string, privacyVer int) (json.RawMessage, error) {
	_, body, err := PostWxJsonWithAuthToken(appid, "/wxa/get_effective_domain", "", struct{}{})
	if err != nil {
		return nil, err
	}
	return json.RawMessage(body), nil
}

type ModifyDomainDirectlyItem struct {
	Action          string   `json:"action" wx:"action" binding:"required"` // add/delete/set/get
	RequestDomain   []string `json:"requestdomain" wx:"requestdomain"`      // request 合法域名；action 为 get 时不需要
	WsRequestDomain []string `json:"wsrequestdomain" wx:"wsrequestdomain"`  // socket 合法域名；action 为 get 时不需要
	UploadDomain    []string `json:"uploaddomain" wx:"uploaddomain"`        // uploadFile 合法域名；action 为 get 时不需要
	DownloadDomain  []string `json:"downloaddomain" wx:"downloaddomain"`    // downloadFile 合法域名；action 为 get 时不需要
	UdpDomain       []string `json:"udpdomain" wx:"udpdomain"`              // udp 合法域名；action 为 get 时不需要
	TcpDomain       []string `json:"tcpdomain" wx:"tcpdomain"`              // tcp 合法域名；action 为 get 时不需要
}

func UpdateModifyDomainDirectly(appid string, req ModifyDomainDirectlyItem) error {
	_, _, err := PostWxJsonWithAuthToken(appid, "/wxa/modify_domain_directly", "", req)
	return err
}
