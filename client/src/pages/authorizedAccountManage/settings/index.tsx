import { Alert, Tabs } from "tdesign-react";
import { useSearchParams } from "react-router-dom";
import PrivacySettings from "./modules/privacy";
import ServerDomainSettings from "./modules/serverDomain";

const { TabPanel } = Tabs;

export default function MiniProgramSettings() {
    const [searchParams] = useSearchParams();
    const appId = searchParams.get("appId") || "";

    if (!appId) {
        return (
            <Alert
                theme="warning"
                message="缺少小程序 AppID，请从代开发小程序列表进入设置页面。"
            />
        );
    }

    return (
        <div>
            <p className="text">小程序设置　/AppID：{appId}</p>
            <Tabs value="privacy" placement="top" size="medium" theme="normal">
                <TabPanel value="privacy" label="用户隐私保护指引">
                    <PrivacySettings appId={appId} />
                </TabPanel>
                <TabPanel value="domain" label="服务器域名">
                    <ServerDomainSettings appId={appId} />
                </TabPanel>
            </Tabs>
        </div>
    );
}
