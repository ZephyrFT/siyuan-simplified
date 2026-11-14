// 模板 HTTP 出站请求已从本版本移除，保留只读占位，避免设置和预览界面发起网络请求。

export type TemplateHTTPPermissionSource = {kind: "file", path: string} |
    {kind: "av", avID: string, keyID: string, slot: "template" | "renderTemplate" | "automation"};

export const templateHTTPPermission = async (action: "get" | "revoke" | "revokeAll",
                                             source?: TemplateHTTPPermissionSource): Promise<boolean | undefined> => {
    void action;
    void source;
    return undefined;
};

export const templateHTTPPermissionLabel = (allowed?: boolean) => {
    void allowed;
    return window.siyuan.languages.templateHTTPNotAllowed;
};

export const createTemplateHTTPPermissionIndicator = (allowed?: boolean) => {
    void allowed;
    const indicator = document.createElement("span");
    indicator.className = "fn__none";
    return indicator;
};

export const templateHTTPPermissionHTML = (slot: "template" | "renderTemplate") => {
    void slot;
    return "";
};

export const bindTemplateHTTPPermission = (element: HTMLElement,
                                          source: Extract<TemplateHTTPPermissionSource, {kind: "av"}>) => {
    void element;
    void source;
};
