import { Button, MessagePlugin, Alert, Textarea } from "tdesign-react";
import { useCallback, useEffect, useState } from "react";
import { request } from "../../../../../utils/axios";
import {
    getEffectiveDomainRequest,
    updateModifyDomainDirectlyRequest,
} from "../../../../../utils/apis";
import { serverDomainConfig } from "../../../../../config/weapp";

type DomainData = Record<string, string>;
const fields: Array<[string, string, string, number, string?]> = [
    ["requestdomain", "request合法域名", "以 https:// 开头", 300, "https://"],
    ["wsrequestdomain", "socket合法域名", "以 wss:// 开头", 200, "wss://"],
    ["uploaddomain", "uploadFile合法域名", "以 https:// 开头", 200, "https://"],
    [
        "downloaddomain",
        "downloadFile合法域名",
        "以 https:// 开头",
        200,
        "https://",
    ],
    ["udpdomain", "udp合法域名", "以 udp:// 开头", 200, "udp://"],
    ["tcpdomain", "tcp合法域名", "以 tcp:// 开头", 200, "tcp://"],
    ["prefetchdomain", "DNS预解析域名", "该域名无需填写协议头", 5],
    [
        "preconnectdomain",
        "预连接域名",
        "以 https:// 开头，且需要为request合法域名与DNS预解析域名的子集",
        5,
        "https://",
    ],
];
const parse = (value?: string) =>
    value
        ?.split(";")
        .map((v) => v.trim())
        .filter(Boolean) || [];
const asDomainList = (value: unknown): string[] => {
    if (Array.isArray(value))
        return value.filter((item) => typeof item === "string");
    if (typeof value === "string") return parse(value);
    return [];
};

export default function ServerDomainSettings({ appId }: { appId: string }) {
    const [domains, setDomains] = useState<DomainData>({});
    const [loading, setLoading] = useState(false);

    const load = useCallback(
        async (showMessagePlugin?: boolean = true) => {
            setLoading(true);
            try {
                const resp: any = await request({
                    request: getEffectiveDomainRequest,
                    data: { appid: appId },
                });
                let data = resp?.data ?? resp;
                if (typeof data === "string") {
                    try {
                        data = JSON.parse(data);
                    } catch {
                        data = {};
                    }
                }
                if (data?.data && typeof data.data === "object")
                    data = data.data;
                // mp_domain、third_domain 等对象；表单应显示生效域名。
                if (data?.effective_domain) data = data.effective_domain;
                const next: DomainData = {};
                fields.forEach(([key]) => {
                    next[key] = asDomainList(data?.[key]).join(";");
                });

                setDomains(next);
                if (showMessagePlugin) {
                    MessagePlugin.success("已刷新生效域名");
                }
            } catch (e: any) {
                MessagePlugin.error(e?.message || "获取生效域名失败");
            } finally {
                setLoading(false);
            }
        },
        [appId],
    );
    useEffect(() => {
        void load(false);
    }, [load]);

    const save = async () => {
        const domainLists = Object.fromEntries(
            fields.map(([key]) => [key, parse(domains[key])]),
        ) as Record<string, string[]>;
        for (const [key, label, , limit, prefix] of fields) {
            const list = domainLists[key];
            if (list.length > limit) {
                MessagePlugin.error(`${label}最多填写 ${limit} 个域名`);
                return;
            }
            if (prefix && list.some((domain) => !domain.startsWith(prefix))) {
                MessagePlugin.error(`${label}必须以 ${prefix} 开头`);
                return;
            }
        }
        const requestDomains = new Set(domainLists.requestdomain);
        const dnsDomains = new Set(domainLists.prefetchdomain);
        const preconnectInvalid = domainLists.preconnectdomain.some(
            (domain) => {
                const hostname = domain.slice("https://".length);
                return !requestDomains.has(domain) || !dnsDomains.has(hostname);
            },
        );
        if (preconnectInvalid) {
            MessagePlugin.error(
                "预连接域名必须同时属于 request 合法域名和 DNS 预解析域名",
            );
            return;
        }
        setLoading(true);
        try {
            await request({
                request: updateModifyDomainDirectlyRequest,
                data: {
                    appid: appId,
                    modify_domain_directly: { action: "set", ...domainLists },
                },
            });
            MessagePlugin.success("服务器域名配置已提交");
            await load();
        } catch (e: any) {
            MessagePlugin.error(e?.message || "快速配置失败");
        } finally {
            setLoading(false);
        }
    };

    const getConfig = () => {
        const serverDomainConfig_: any = serverDomainConfig;
        const next: DomainData = {};
        fields.forEach(([key]) => {
            next[key] = asDomainList(serverDomainConfig_?.[key]).join(";");
        });

        console.log(fields, "fields", next);
        setDomains(next);
    };

    return (
        <div style={{ maxWidth: 900 }}>
            <Alert
                style={{ marginTop: 8 }}
                theme="info"
                message="域名之间请用 ; 分割。"
            />
            <div
                style={{
                    width: "100%",
                    display: "flex",
                    flexDirection: "column",
                }}
            >
                {fields.map(([key, label, hint, limit]) => (
                    <div key={key} style={{ marginTop: 16 }}>
                        <div style={{ marginBottom: 6 }}>{label}</div>
                        <Textarea
                            value={domains[key] || ""}
                            onChange={(value) =>
                                setDomains((v: any) => ({
                                    ...v,
                                    [key]: value,
                                }))
                            }
                            placeholder={`${hint}。最多填写${limit}个域名，域名间请用 ; 分割`}
                        />
                    </div>
                ))}
                <div style={{ marginTop: 20, display: "flex", gap: 12 }}>
                    <Button theme="primary" loading={loading} onClick={save}>
                        保存配置
                    </Button>
                    <Button loading={loading} onClick={() => void load()}>
                        刷新生效域名
                    </Button>
                    <Button
                        variant="outline"
                        loading={loading}
                        onClick={getConfig}
                    >
                        一键默认配置
                    </Button>
                </div>
            </div>
        </div>
    );
}
