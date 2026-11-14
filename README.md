# SiYuan Simplified（思源简版）

本地优先、去 AI、去云端的 SiYuan 分支。保留块编辑器、双链、搜索、导出、加密笔记本和本地快照，移除账户、云服务、集市、插件运行时、AI/Agent、在线更新等功能。

> 本仓库是 [siyuan-note/siyuan](https://github.com/siyuan-note/siyuan) 的非官方修改版，与云南理智科技有限公司无隶属关系。修改内容按 AGPL-3.0-or-later 发布，详见[许可与上游](#许可与上游)。

[Releases](https://github.com/ZephyrFT/siyuan-simplified/releases) · [问题反馈](https://github.com/ZephyrFT/siyuan-simplified/issues)

---

## 这是什么

SiYuan Simplified 面向只需要“本地笔记”的场景：数据留在自己的工作空间，应用不要求登录，不连接官方账户或云服务，也不内置 AI 与插件市场。除了 PlantUML 的在线渲染外，本分支不主动访问外部网络。

它沿用 SiYuan 的工作空间、`.sy` 文档格式和块级数据模型，因此可以打开标准 SiYuan 工作空间；但被移除的功能不会恢复。

## 与上游的主要差异

### 已移除

- **AI 与智能体**：AI 对话、编辑器内 AI、AI 记忆、技能、思考卡、AI OCR、模型供应商配置，以及 ChatGPT、Claude、Gemini、DeepSeek 等接入。
- **Agent 与 MCP**：Agent 运行时、Agent 聊天侧栏、MCP 服务与 OAuth 相关接口。
- **集市与插件运行时**：插件/主题/图标/模板市场、安装与更新、评分、发布，以及内核侧第三方插件加载。前端仍保留部分插件 API 类型声明以维持编译兼容，但不再加载第三方插件。
- **账户与云服务**：登录、账户界面、云用户、订阅与公告刷新、云同步入口及相关设置页。
- **在线能力**：联网搜索、网页抓取、embedding/rerank、云端 OCR、在线更新检查与内置更新器、elevator 提权组件。
- **界面模块**：收集箱（Inbox）、Agent 聊天栏、AI/集市/账户相关设置页与移动端页面。
- **其他**：随应用分发的变更日志模块、旧的在线发布辅助逻辑。

### 保留

- 块编辑器、所见即所得 Markdown、双链与反链、大纲、标签、书签。
- 数据库/属性视图、模板、闪卡、全局搜索。
- 导出（PDF、HTML、Markdown、图片等）、导入、资源文件管理。
- 加密笔记本、本地快照与历史、数据仓库检查。
- 本地 OCR 运行时与本地模型。
- 设置、快捷键、外观主题、字体与代码片段。
- PlantUML 在线渲染（见下）。

### 唯一保留的外联

PlantUML 仍然使用在线渲染服务。除 PlantUML 外，本分支不包含主动访问外部网络的业务功能；如果你需要完全离线，可以在设置中避免使用 PlantUML 图表。

## 与官方版共存

本分支使用独立的品牌、安装标识和配置目录，可以和官方 SiYuan 同时安装：

| 项目 | 值 |
|---|---|
| 显示名称 | SiYuan Simplified / 思源简版 |
| 应用标识 | `org.zephyrft.siyuan-simplified` |
| 用户配置 | `~/.config/siyuan-simplified` |
| Electron 数据 | `%APPDATA%\siyuan-simplified-Electron` |
| 默认工作空间 | `~/SiYuan-Simplified` |
| 安装包 | `siyuan-simplified-${version}-${os}.${ext}` |

以下内部名称刻意保留，避免破坏数据格式与前端兼容：`window.siyuan`、`SiYuan-Kernel(.exe)`、`.sy`、`siyuan://` 协议、`X-SiYuan-App-ID`。

## 下载与安装

从 [Releases](https://github.com/ZephyrFT/siyuan-simplified/releases) 下载 Windows 安装包。安装包未做代码签名，Windows SmartScreen 可能提示“未知发布者”，请核对仓库与 SHA256 后再运行。

首次测试建议：

1. 退出正在运行的 SiYuan / SiYuan Simplified。
2. 备份笔记，或复制一份测试工作空间：
   ```bat
   robocopy "D:\笔记" "D:\笔记-test" /E
   ```
3. 使用全新配置目录启动，并打开测试副本：
   ```bat
   "SiYuan Simplified.exe" --workspace="D:\笔记-test"
   ```

本分支的配置目录与官方版不同，不会修改官方版的 `~/.config/siyuan`。但被移除功能相关的数据（AI 会话、集市包等）不会被新版本使用。

## 从源码构建

### 依赖

| 依赖 | 版本/说明 |
|---|---|
| Go | 见 `kernel/go.mod`（当前 1.26.x） |
| Node.js | 24.x，建议 24.22 或更新（测试模块 mock 需要） |
| pnpm | 12.6.0 |
| MinGW-w64 GCC | Windows 内核 CGO 编译 |
| Python | 3.11+，打包 OCR 资源用 |
| Visual C++ Redistributable | 打包 Windows OCR 运行时用 |

本机可把工具放在任意目录；作者环境放在 `D:\Library\siyuan-tools`：

```text
D:\Library\siyuan-tools\nodejs-24.22.3
D:\Library\siyuan-tools\pnpm-12.6.0
D:\Library\siyuan-tools\w64devkit
```

### 构建

```bash
# 1. 前端依赖与产物
cd app
pnpm install --frozen-lockfile
pnpm run build

# 2. Windows 内核
cd ../kernel
CGO_ENABLED=1 CC=gcc GOOS=windows GOARCH=amd64 \
  go build -tags "fts5 sqlcipher" -ldflags "-s -w" \
  -o ../app/kernel/SiYuan-Kernel.exe .

# 3. 安装包
cd ../app
pnpm exec electron-builder --config electron-builder.yml --publish=never
```

产物位于 `app/build/`。安装包的 OCR 资源由 `scripts/prepare-ocr.py` 在打包前准备，Electron 主进程由 `scripts/build-electron.mjs` 编译。

### 自动发布

`.github/workflows/release-windows.yml` 会在推送 `v*` 标签时构建 Windows 安装包并创建 GitHub Release。标签需要与 `app/package.json` 的版本一致：

```bash
git tag v3.8.10-simplified.1
git push origin v3.8.10-simplified.1
```

工作流会下载依赖、准备 OCR 资源、编译前端与内核，然后上传 `siyuan-simplified-*-win.exe`。

## 测试

```bash
# 前端
cd app
pnpm test

# 内核（Windows 需要 MinGW）
cd ../kernel
CGO_ENABLED=1 CC=gcc go test -tags "fts5 sqlcipher" ./...
```

前端测试需要 Node 24.22+；24.13 的 `--experimental-test-module-mocks` 对 CommonJS 有缺陷，会导致 Electron 替身测试失败。

## 数据兼容

- 文档仍使用标准 `.sy` 格式，工作空间结构未变。
- 加密笔记本、历史、快照和资源文件的读取逻辑保留。
- 新版本使用独立配置目录，不会自动继承官方版的设置；需要时手动迁移或在首次启动时重新选择工作空间。
- 升级前请备份工作空间，尤其是只在一个版本中测试过的新功能。

## 项目结构

| 路径 | 说明 |
|---|---|
| `kernel/` | Go 内核：块树、索引、搜索、导出、加密、本地服务 |
| `app/` | Electron + TypeScript 前端，`app/src/protyle/` 为编辑器 |
| `app/electron/` | 主进程、窗口、内核启动与 IPC |
| `app/guide/` | 内置指南（四种语言） |
| `scripts/` | 打包、OCR 资源、发布校验脚本 |

## 许可与上游

- 许可证：[AGPL-3.0-or-later](LICENSE)，保留原始版权与 `THIRD_PARTY_NOTICES.md`。
- 上游：https://github.com/siyuan-note/siyuan
- 本仓库为修改版，名称、图标、安装标识和配置目录均已调整；发布衍生版本时请保留 AGPL 源码可获取性与上游署名。
- 其他语言的 `README.*.md` 文件继承自上游，可能与当前分支不一致；以本文件为准。

---

## English summary

SiYuan Simplified is an unofficial, local-first fork of SiYuan. It keeps the block editor, backlinks, search, export, encrypted notebooks, local snapshots and local OCR, while removing AI/agent/MCP features, the bazaar and plugin runtime, accounts and cloud services, online search, and the built-in updater. PlantUML online rendering is the only remaining outbound feature. It uses a separate app ID and config directory, so it can be installed alongside official SiYuan. Licensed under AGPL-3.0-or-later; see the upstream project for attribution.
