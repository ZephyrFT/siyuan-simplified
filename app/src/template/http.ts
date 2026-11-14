// 模板 HTTP 出站请求已从本版本移除，保留无网络守卫接口供渲染路径兼容调用。

export const cancelTemplateHTTPExecutions = () => {
};

export const getTemplateHTTPWindow = (): {app: string; session: string} | undefined => undefined;

export const revokeTemplateHTTPConfirmation = (token: string, owner: {app: string; session: string}) => {
    void token;
    void owner;
};

export const cancelTemplateHTTPForElement = (element: Element,
                                             options?: {preserveAVMetadata?: boolean; preserveManualAutomation?: boolean}) => {
    void element;
    void options;
};

// 始终返回可中止的本地信号，不再等待内核确认或发送请求。
export const guardTemplateHTTPRequest = (isCurrent: () => boolean, owner?: Element) => {
    void isCurrent;
    void owner;
    const controller = new AbortController();
    return {signal: controller.signal, cancel: () => controller.abort(), dispose: () => {
        controller.abort();
    }};
};

export const beginTemplateHTTPRequest = (url: string,
                                         data: {id?: string; blockID?: string; ignoreRows?: boolean; rows?: boolean},
                                         headers?: Record<string, string>,
                                         signal?: AbortSignal): undefined => {
    void url;
    void data;
    void headers;
    void signal;
    return undefined;
};

export const processTemplateHTTPEvent = (cmd: string, data: unknown) => {
    void cmd;
    void data;
};
