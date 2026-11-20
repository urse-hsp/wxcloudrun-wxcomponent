export const userPrivacyConfig = {
    owner_setting: {
        contact_email: "eason@dongya58.com",
        notice_method: "打电话",
    },
    setting_list: [
        {
            privacy_key: "UserInfo",
            privacy_text: "用于报名活动与订场",
            privacy_label: "用户信息（微信昵称、头像）",
        },
        {
            privacy_key: "PhoneNumber",
            privacy_text: "快速登录小程序",
            privacy_label: "手机号",
        },
        {
            privacy_key: "Clipboard",
            privacy_text: "剪切板",
            privacy_label: "剪切板",
        },
        {
            privacy_key: "ChooseLocation",
            privacy_text: "选择的位置信息",
            privacy_label: "选择的位置信息",
        },
        {
            privacy_key: "EXUserPublishContent",
            privacy_text: "发布内容",
            privacy_label: "发布内容",
        },
        {
            privacy_key: "Location",
            privacy_text: "选择的位置信息",
            privacy_label: "位置信息",
        },
    ],
};

export const serverDomainConfig = {
    requestdomain: [
        "https://api-pre.dongya58.com",
        "https://api-pre.wanli58.com",
        "https://api.dongya58.com",
        "https://api.wanli58.com",
    ],
    uploaddomain: ["https://up-z0.qiniup.com", "https://up-z1.qiniup.com"],
    downloaddomain: [
        "https://pic.dongya58.com",
        "https://pic.wanli58.com",
        "https://resource.wanli58.com",
        "https://wanlires.dongya58.com",
        "https://wanlires.wanli58.com",
    ],
};
