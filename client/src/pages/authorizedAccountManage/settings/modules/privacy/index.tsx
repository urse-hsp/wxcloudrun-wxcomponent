import {
    Alert,
    Button,
    Checkbox,
    Input,
    MessagePlugin,
    Textarea,
} from "tdesign-react";
import { useEffect, useMemo, useState } from "react";
import { request } from "../../../../../utils/axios";
import {
    getUserPrivacyRequest,
    updateUserPrivacyRequest,
} from "../../../../../utils/apis";
import { userPrivacyConfig } from "../../../../../config/weapp";

type IOwnerSetting = {
    contact_email: string;
    contact_phone: string;
    contact_qq: string;
    contact_weixin: string;
    ext_file_media_id: string;
    notice_method: string;
    store_expire_timestamp: string;
};

type IPrivacyItem = {
    privacy_key: string;
    privacy_text: string;
    privacy_label?: string;
};

type IPrivacyDesc = {
    privacy_key: string;
    privacy_desc: string;
};

const emptyOwnerSetting: IOwnerSetting = {
    contact_email: "",
    contact_phone: "",
    contact_qq: "",
    contact_weixin: "",
    ext_file_media_id: "",
    notice_method: "",
    store_expire_timestamp: "",
};

type Props = { appId: string };

/**
 * Normalize the default privacy items declared in config/weapp.ts.
 * Config entries may only declare privacy_key / privacy_desc / privacy_label,
 * while the editable list needs privacy_text, so fall back in that order.
 */
const normalizeDefaultSettingList = (
    list: Array<Partial<IPrivacyItem> & { privacy_desc?: string }> = [],
): IPrivacyItem[] =>
    list
        .filter((item) => Boolean(item?.privacy_key))
        .map((item) => ({
            privacy_key: item.privacy_key as string,
            privacy_text:
                item.privacy_text || item.privacy_desc || item.privacy_label || "",
            privacy_label: item.privacy_label,
        }));

export default function PrivacySettings({ appId }: Props) {
    const [loading, setLoading] = useState(false);
    const [saving, setSaving] = useState(false);
    const [ownerSetting, setOwnerSetting] =
        useState<IOwnerSetting>(emptyOwnerSetting);
    const [availableItems, setAvailableItems] = useState<IPrivacyDesc[]>([]);
    const [settingList, setSettingList] = useState<IPrivacyItem[]>([]);

    const selectedKeys = useMemo(
        () => new Set(settingList.map((item) => item.privacy_key)),
        [settingList],
    );

    useEffect(() => {
        if (appId) loadPrivacySetting();
    }, [appId]);

    const loadPrivacySetting = async (setting_list?: IPrivacyItem[]) => {
        setLoading(true);
        const resp: any = await request({
            request: getUserPrivacyRequest,
            data: { appid: appId, privacy_ver: 2 },
        });
        if (resp.code === 0) {
            const data = resp.data || {};
            setsSetting_list({
                ...data,
                setting_list:
                    setting_list && setting_list.length > 0
                        ? setting_list
                        : data?.setting_list,
            });
            setOwnerSetting({
                ...emptyOwnerSetting,
                ...(data.owner_setting || {}),
            });
        }
        setLoading(false);
    };

    const setsSetting_list = (data: any) => {
        const descList =
            data.privacy_desc?.privacy_desc_list ||
            data.privacy_desc_list ||
            [];
        const existingList: IPrivacyItem[] = data.setting_list || [];
        const choices: IPrivacyDesc[] = [...descList];
        existingList.forEach((item) => {
            if (
                !choices.some(
                    (choice) => choice.privacy_key === item.privacy_key,
                )
            ) {
                choices.push({
                    privacy_key: item.privacy_key,
                    privacy_desc: item.privacy_label || item.privacy_key,
                });
            }
        });
        (data.privacy_list || []).forEach((key: string) => {
            if (!choices.some((choice) => choice.privacy_key === key)) {
                choices.push({ privacy_key: key, privacy_desc: key });
            }
        });
        // Keep the option list already fetched from the server, so applying the
        // config defaults never hides the other selectable privacy items.
        setAvailableItems((previous) => {
            const merged = [...choices];
            previous.forEach((item) => {
                if (
                    !merged.some(
                        (choice) => choice.privacy_key === item.privacy_key,
                    )
                ) {
                    merged.push(item);
                }
            });
            return merged;
        });
        setSettingList(
            existingList.map((item) => ({
                privacy_key: item.privacy_key,
                privacy_text: item.privacy_text || "",
                privacy_label: item.privacy_label,
            })),
        );
    };

    const getConfig = () => {
        const config: any = userPrivacyConfig || {};
        const setting_list = normalizeDefaultSettingList(config.setting_list);

        setOwnerSetting({
            ...emptyOwnerSetting,
            ...(config.owner_setting || {}),
        });

        // Fill the whole form (available items + checked items) with the
        // defaults declared in config/weapp.ts.
        setsSetting_list({
            ...config,
            setting_list,
        });
    };

    const togglePrivacyItem = (item: IPrivacyDesc, checked: boolean) => {
        if (checked) {
            setSettingList((current) =>
                current.some((value) => value.privacy_key === item.privacy_key)
                    ? current
                    : [
                          ...current,
                          {
                              privacy_key: item.privacy_key,
                              privacy_text: item.privacy_desc || "",
                          },
                      ],
            );
        } else {
            setSettingList((current) =>
                current.filter(
                    (value) => value.privacy_key !== item.privacy_key,
                ),
            );
        }
    };

    const updatePrivacyText = (key: string, text: string) => {
        setSettingList((current) =>
            current.map((item) =>
                item.privacy_key === key
                    ? { ...item, privacy_text: text }
                    : item,
            ),
        );
    };

    const savePrivacySetting = async () => {
        setSaving(true);
        const storeExpireTimestamp = ownerSetting.store_expire_timestamp.trim();
        const resp = await request({
            request: updateUserPrivacyRequest,
            data: {
                appid: appId,
                privacy_ver: 2,
                owner_setting: {
                    ...ownerSetting,
                    // This field is a duration such as "30天", not a Unix timestamp.
                    store_expire_timestamp:
                        storeExpireTimestamp === "0"
                            ? ""
                            : storeExpireTimestamp,
                },
                setting_list: settingList.map(
                    ({ privacy_key, privacy_text }) => ({
                        privacy_key,
                        privacy_text,
                    }),
                ),
            },
        });
        if (resp.code === 0) {
            MessagePlugin.success("隐私保护指引设置已提交");
            loadPrivacySetting();
        }
        setSaving(false);
    };

    const updateOwnerSetting = (key: keyof IOwnerSetting, value: string) => {
        setOwnerSetting((current) => ({ ...current, [key]: value }));
    };

    return (
        <div style={{ padding: "16px 0" }}>
            {/* <Alert
                theme="info"
                message="当前设置的是开发版隐私保护指引（privacy_ver=2）。请填写开发者信息，并选择小程序实际使用的隐私权限。"
            /> */}
            <div style={{ marginTop: 20 }}>
                <p className="text">开发者信息</p>
                <div
                    style={{
                        display: "grid",
                        gridTemplateColumns: "repeat(2, minmax(280px, 1fr))",
                        gap: 16,
                        maxWidth: 900,
                    }}
                >
                    <label>
                        联系邮箱
                        <Input
                            value={ownerSetting.contact_email}
                            onChange={(value: any) =>
                                updateOwnerSetting("contact_email", value)
                            }
                        />
                    </label>
                    <label>
                        联系电话
                        <Input
                            value={ownerSetting.contact_phone}
                            onChange={(value: any) =>
                                updateOwnerSetting("contact_phone", value)
                            }
                        />
                    </label>
                    <label>
                        联系 QQ
                        <Input
                            value={ownerSetting.contact_qq}
                            onChange={(value: any) =>
                                updateOwnerSetting("contact_qq", value)
                            }
                        />
                    </label>
                    <label>
                        联系微信
                        <Input
                            value={ownerSetting.contact_weixin}
                            onChange={(value: any) =>
                                updateOwnerSetting("contact_weixin", value)
                            }
                        />
                    </label>
                    <label>
                        通知方式
                        <Input
                            value={ownerSetting.notice_method}
                            onChange={(value: any) =>
                                updateOwnerSetting("notice_method", value)
                            }
                        />
                    </label>
                    <label>
                        数据留存期限（留空或填写“数字+天”，如 30天）
                        <Input
                            value={ownerSetting.store_expire_timestamp}
                            onChange={(value: any) =>
                                updateOwnerSetting(
                                    "store_expire_timestamp",
                                    value,
                                )
                            }
                            placeholder="留空使用默认说明，例如：30天"
                        />
                    </label>
                    <label style={{ gridColumn: "1 / -1" }}>
                        隐私指引文件 MediaID
                        <Input
                            value={ownerSetting.ext_file_media_id}
                            onChange={(value: any) =>
                                updateOwnerSetting("ext_file_media_id", value)
                            }
                        />
                    </label>
                </div>
            </div>

            <div style={{ marginTop: 24 }}>
                <p className="text">隐私权限及说明</p>
                {availableItems.length === 0 && !loading && (
                    <Alert
                        theme="warning"
                        message="微信接口未返回可选权限项。请检查该小程序的授权状态和微信接口返回内容。"
                    />
                )}
                {availableItems.map((item) => {
                    const configured = settingList.find(
                        (value) => value.privacy_key === item.privacy_key,
                    );
                    return (
                        <div
                            key={item.privacy_key}
                            style={{
                                maxWidth: 900,
                                padding: "12px 0",
                                borderBottom: "1px solid #eee",
                            }}
                        >
                            <Checkbox
                                checked={selectedKeys.has(item.privacy_key)}
                                onChange={(checked) =>
                                    togglePrivacyItem(item, checked)
                                }
                            >
                                {item.privacy_desc || item.privacy_key}（
                                {item.privacy_key}）
                            </Checkbox>
                            {configured && (
                                <Textarea
                                    value={configured.privacy_text}
                                    onChange={(value: any) =>
                                        updatePrivacyText(
                                            item.privacy_key,
                                            value,
                                        )
                                    }
                                    placeholder="直接填写收集用途，无需添加“用于”或“为了”"
                                    style={{
                                        margin: "8px 0 0 26px",
                                        width: "calc(100% - 26px)",
                                    }}
                                />
                            )}
                        </div>
                    );
                })}
            </div>
            <div style={{ marginTop: 20 }}>
                <Button
                    theme="primary"
                    loading={saving}
                    disabled={loading}
                    onClick={savePrivacySetting}
                >
                    保存设置
                </Button>
                <Button
                    variant="outline"
                    style={{ marginLeft: 12 }}
                    loading={loading}
                    onClick={() => loadPrivacySetting()}
                >
                    重新获取
                </Button>
                <Button
                    variant="outline"
                    style={{ marginLeft: 12 }}
                    loading={loading}
                    onClick={getConfig}
                >
                    一键默认配置
                </Button>
            </div>
        </div>
    );
}