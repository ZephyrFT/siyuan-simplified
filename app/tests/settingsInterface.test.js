const assert = require("node:assert/strict");
const {readFileSync} = require("node:fs");
const {test} = require("node:test");
const {runInNewContext} = require("node:vm");
const ts = require("typescript");
const preprocess = require("ifdef-loader/preprocessor").parse;

const setup = ({mobile = false, browser = false, settingsWindow = false, native = false} = {}) => {
    const calls = [];
    const app = {};
    const window = {siyuan: {dialogs: [], config: {system: {settingsWindow: native}, bazaar: {trust: true}},
        languages: {config: "Settings"}}};
    const selector = '.config__tab-container[data-name="bazaar"]';
    const element = () => {
        const attributes = new Map();
        const content = {};
        const container = {style: {}};
        return {setAttribute: (key, value) => attributes.set(key, value), getAttribute: key => attributes.get(key),
            querySelector: query => query === ".config__tab-container" ? {} :
                query === selector ? content : query === ".b3-dialog__container" ? container : null,
            querySelectorAll: () => []};
    };
    class Dialog {
        constructor(options) {
            this.element = element();
            this.options = options;
            window.siyuan.dialogs.push(this);
            calls.push(["dialog"]);
        }
        destroy() {
            window.siyuan.dialogs.splice(window.siyuan.dialogs.indexOf(this), 1);
            calls.push(["destroy", this]);
            this.options.destroyCallback?.();
        }
    }
    const constants = {DIALOG_SETTING: "settings"};
    const modules = {
        "../mobile/menu": {openMobileSetting: (...args) => calls.push(["mobile", ...args])},
        "./search/dialog": {initSettingSearch() {}, switchSettingTab: (_element, _app, tab) => calls.push(["tab", tab])},
        "./setting/save": {bindSettingSaveDelegation() {}},
        "../dialog": {Dialog}, "../constants": {Constants: constants},
        "../protyle/util/selectionOffsets": {}, "./setting/drag": {initSettingDrag: () => () => calls.push(["disposeDrag"])},
        "./bazaarTab": {unmountBazaarTab: () => calls.push(["unmountBazaar"]),
            withMountedBazaar: async callback => callback({bazaar: {switchBazaarTab: (...args) => calls.push(["bazaar", ...args])},
                renderReadme: (...args) => calls.push(["readme", ...args])})},
        "../util/fetch": {fetchSyncPost: async () => ({code: 0, data: {packages: [{name: "example"}]}})},
        "../util/functions": {getFrontend: () => mobile ? "mobile" : "desktop"},
        "../dialog/message": {}, "../util/escape": {escapeHtml: value => value, escapeHtmlTextAndAttr: value => value},
        "../util/bazaarAvailability": {isBazaarAvailable: () => true},
        "./setting/tabs": {getSettingTabDefs: () => []},
        "./tabs/accessRuntime": {clearAccessTabElement() {}}, "./tabs/syncRuntime": {clearSyncTabElement() {}},
        "./tabs/ai/aiProviderUi": {openChatGPTProvider: () => calls.push(["chatgpt"])},
        "./assets": {unmountAssetsTab() {}}, "./ocr": {unmountOCRTab() {}},
        "../util/hostCapabilities": {getHostCapabilities: () => ({documentImportExport: true})},
        "./tabs/workspaceStorage": {unmountWorkspaceStorage() {}},
        "./setting/nativeWindow": {
            openNativeSettings: async (...args) => calls.push(["native", ...args]),
        },
        "./setting/interface": {prepareSettingDialog: () => calls.push(["prepareDialog"]),
            closeSettingDialog: () => window.siyuan.dialogs.find(item =>
                item.element.querySelector(".config__tab-container"))?.destroy()},
        "./setting/windowContext": {isSettingsWindow: () => settingsWindow},
        "./setting/windowDialog": {fitSettingsWindowDialog: () => calls.push(["fit"])},
    };
    window.close = () => calls.push(["closeSelf"]);
    const source = preprocess(readFileSync("src/config/index.ts", "utf8"), {MOBILE: mobile, BROWSER: browser});
    const code = ts.transpileModule(source, {
        compilerOptions: {module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2021},
    }).outputText;
    const exports = {};
    runInNewContext(code, {exports, window, getSelection: () => ({rangeCount: 0}), require: name => {
        assert.ok(modules[name], name);
        return modules[name];
    }});
    return {api: exports, app, calls, window, Dialog, constants};
};

test("opening a settings dialog closes the builtin native window before mounting", () => {
    const {api, app, calls} = setup();
    api.openSetting(app, "app");
    assert.deepEqual(calls.slice(0, 2), [["prepareDialog"], ["dialog"]]);
});

test("the plugin settings API also closes the builtin native window", () => {
    const {api, app, calls} = setup({native: true});
    api.openPluginSetting(app);
    assert.deepEqual(calls.slice(0, 2), [["prepareDialog"], ["dialog"]]);
    assert.equal(calls.some(call => call[0] === "native"), false);
});

test("reopening dialog settings cleans up the existing dialog", () => {
    const {api, app, calls, window} = setup();
    const first = api.openSetting(app);
    const second = api.openSetting(app);
    assert.notEqual(second, first);
    assert.equal(window.siyuan.dialogs.length, 1);
    assert.equal(window.siyuan.dialogs[0], second);
    assert.equal(calls.filter(call => call[0] === "destroy").length, 1);
});

for (const mobile of [false, true]) {
    test(`browser and mobile settings retain their opening paths (mobile=${mobile})`, () => {
        const {api, app, calls} = setup({mobile, browser: true, native: true});
        api.openSetting(app, "app");
        assert.equal(calls.some(call => call[0] === "native" || call[0] === "prepareDialog"), false);
        assert.equal(calls[0][0], mobile ? "mobile" : "dialog");
    });
}
