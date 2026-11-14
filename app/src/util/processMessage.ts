/// #if !MOBILE
import {reloadUI} from "../layout/util";
import {closeSettingsWindow} from "../config/setting/windowContext";
/// #else
import {reloadMobileUI} from "../mobile/util/reloadUI";
/// #endif
import {hideMessage, showMessage} from "../dialog/message";
import {isBrowser} from "./functions";
import {applySettingTask} from "../config/setting/taskBlocker";
import {notifyOCRChanged} from "../config/ocrRuntime";

export const processMessage = (response: IWebSocketData) => {
    if (response.cmd === "notebookIconChanged") {
        const {updateNotebookBreadcrumbIcon}: typeof import("../protyle/breadcrumb/iconRefresh") = require("../protyle/breadcrumb/iconRefresh");
        updateNotebookBreadcrumbIcon(response.data.boxID, response.data.icon);
    }
    if (response.cmd === "closeBox" || response.cmd === "removeBox") {
        const {invalidateRelatedAVSecurity}: typeof import("../protyle/render/av/relatedView") = require("../protyle/render/av/relatedView");
        invalidateRelatedAVSecurity();
    }
    if (response.cmd === "assetContentIndexCompleted" && response.code === 0) {
        const {refreshAssetContentSearch}: typeof import("../search/assets") = require("../search/assets");
        refreshAssetContentSearch();
        return false;
    }
    if (response.cmd === "ocrChanged") {
        notifyOCRChanged();
        return false;
    }
    if (response.cmd === "prepareSettingsReset") {
        void import("../config/setting/reset").then(reset => reset.prepareSettingsReset(response.data));
        return false;
    }
    if (response.cmd === "cancelSettingsReset") {
        void import("../config/setting/reset").then(reset => reset.cancelSettingsReset(response.data));
        return false;
    }
    if (response.cmd === "settingsReset") {
        void import("../config/setting/reset").then(reset => reset.completeSettingsReset(response.data));
        return false;
    }
    if (response.cmd === "settingTask") {
        applySettingTask(response.data);
        return false;
    }
    if ("msg" === response.cmd) {
        showMessage(response.msg, response.data.closeTimeout, response.code === 0 ? "info" : "error", response.data.id);
        return false;
    }
    if ("cmsg" === response.cmd) {
        hideMessage(response.data.id);
        return false;
    }
    if ("cprogress" === response.cmd) {
        const progressElement = document.getElementById("progress");
        if (progressElement) {
            progressElement.remove();
        }
        return false;
    }
    if ("reloadui" === response.cmd) {
        /// #if !MOBILE
        if (closeSettingsWindow()) return false;
        /// #endif
        /// #if MOBILE
        void reloadMobileUI(!!response.data?.resetScroll);
        /// #else
        void reloadUI(!!response.data?.resetScroll);
        /// #endif
        return false;
    }
    if ("reloadpublishpage" === response.cmd) {
        if (window.siyuan.isPublish) {
            window.location.reload();
        }
        return false;
    }
    if ("closepublishpage" === response.cmd) {
        handlePublishServiceClosed(response.msg);
        return false;
    }

    // 小于 0 为提示：-2 提示；-1 报错，大于 0 的错误需处理，等于 0 的为正常操作
    if (response.code < 0) {
        showMessage(response.msg, response.data ? (response.data.closeTimeout || 0) : 0, response.code === -1 ? "error" : "info");
        return false;
    }

    return response;
};

export const handlePublishServiceClosed = (msg: string) => {
    if (isBrowser()) {
        sessionStorage.setItem("siyuanPublishServiceClosed", msg || "");
        window.location.reload();
    }
};

export const checkPublishServiceClosed = (): boolean => {
    if (isBrowser()) {
        const publishServiceClosedMsg = sessionStorage.getItem("siyuanPublishServiceClosed");
        if (publishServiceClosedMsg) {
            sessionStorage.removeItem("siyuanPublishServiceClosed");
            document.body.innerHTML = `<div style="display:flex;align-items:center;justify-content:center;height:100vh">${publishServiceClosedMsg}</div>`;
            return true;
        }
    }
    return false;
};
