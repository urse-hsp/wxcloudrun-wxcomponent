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
