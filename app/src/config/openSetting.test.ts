import * as assert from "node:assert/strict";
import {readFileSync} from "node:fs";
import {resolve} from "node:path";
import {test} from "node:test";
import {runInNewContext} from "node:vm";
import {createSourceFile, isVariableStatement, ModuleKind, ScriptTarget, transpileModule} from "typescript";

const {parse} = require("ifdef-loader/preprocessor");

// openSetting / openBazaarReadme 按工作空间偏好决定默认打开位置：独立窗口或主窗口内对话框
const fixture = (browser = false, mobile = false) => {
    const calls: unknown[][] = [];
    const system: Record<string, unknown> = {settingsWindow: false};
    const config = {bazaar: {trust: true}, system};
    const index = createSourceFile("index.ts", parse(readFileSync(resolve(process.cwd(), "src/config/index.ts"), "utf8"),
        {BROWSER: browser, MOBILE: mobile}, false, true), ScriptTarget.ES2021, true);
    const declarations = index.statements.filter(statement => isVariableStatement(statement) &&
        statement.declarationList.declarations.some(item => ["openSetting", "openBazaarReadme"].includes(item.name.getText(index))));
    const entry = {} as {openSetting: (app: unknown, tab?: string, options?: {aiProvider?: "chatgpt"}) => void;
        openBazaarReadme: (app: unknown, type: string, name: string, from: string) => Promise<void>};
    let native = false;
    const root = {};
    const dialog = {element: {getAttribute: () => "settings", querySelector: () => root}};
    runInNewContext(transpileModule(declarations.map(item => item.getText(index)).join("\n"),
        {compilerOptions: {module: ModuleKind.CommonJS, target: ScriptTarget.ES2021}}).outputText, {
        exports: entry, isSettingsWindow: () => native,
        isBazaarAvailable: () => true, getHostCapabilities: () => ({documentImportExport: true}),
        openNativeSettings: (...args: unknown[]) => calls.push(["native", ...args]),
        closeSettingDialog() {},
        openSettingDialog: (...args: unknown[]) => { calls.push(["dialog", ...args]); return dialog; },
        openMobileSetting: (...args: unknown[]) => calls.push(["mobile", ...args]),
        openChatGPTProvider: () => calls.push(["chatgpt"]),
        window: {siyuan: {dialogs: [dialog], config}},
        Constants: {DIALOG_SETTING: "settings"}, switchSettingTab: (...args: unknown[]) => calls.push(["switch", ...args]),
        fetchSyncPost: async () => ({code: 0, data: {packages: [{name: "plugin"}]}}),
        getFrontend: () => "desktop", withMountedBazaar: async () => calls.push(["readme"]),
    });
    return {calls, entry,
        setPreference: (value: unknown) => { system.settingsWindow = value; },
        setNative: () => { native = true; }};
};

test("only a truthy workspace preference opens the independent settings window", () => {
    for (const [stored, expected] of [[false, "dialog"], [0, "dialog"], ["", "dialog"],
        [true, "native"], [1, "native"]] as const) {
        const f = fixture();
        f.setPreference(stored);
        f.entry.openSetting({});
        assert.equal(f.calls[0][0], expected, `stored ${JSON.stringify(stored)}`);
    }
});

test("browser and mobile keep their own settings UI regardless of the native preference", () => {
    for (const [browser, mobile, expected] of [[true, false, "dialog"], [true, true, "mobile"]] as const) {
        const f = fixture(browser, mobile);
        f.setPreference(true);
        f.entry.openSetting({});
        assert.equal(f.calls[0][0], expected);
    }
});
