#!/usr/bin/env node
// 本地生成 release 安装包：前端产物、Windows 内核、OCR 资源、NSIS 安装包与 SHA256。
// GitHub Actions 不可用时，可用 --publish 和已登录的 gh 直接创建 GitHub Release。
import {createHash} from "node:crypto";
import {spawnSync} from "node:child_process";
import {existsSync, mkdirSync, readdirSync, readFileSync, rmSync, statSync, writeFileSync} from "node:fs";
import path from "node:path";
import {fileURLToPath} from "node:url";

const root = path.resolve(fileURLToPath(new URL("..", import.meta.url)));
const appDir = path.join(root, "app");
const kernelDir = path.join(root, "kernel");
const args = process.argv.slice(2);

const readArg = name => {
    const inline = args.find(item => item.startsWith(name + "="));
    if (inline) return inline.slice(name.length + 1);
    const index = args.indexOf(name);
    return 0 <= index ? args[index + 1] : undefined;
};
const has = name => args.includes(name);
const dryRun = has("--dry-run");
const skipInstall = has("--skip-install");
const publish = has("--publish");
const prerelease = has("--prerelease");
const tag = readArg("--tag");

if (has("--help") || has("-h")) {
    console.log(`用法: pnpm release:local [选项]

默认仅构建 Windows release 安装包，不访问网络发布。

选项:
  --skip-install    跳过 pnpm install
  --dry-run         只打印将执行的步骤
  --publish         构建后用 gh 创建/更新 GitHub Release（需要 --tag）
  --tag <tag>       发布标签，例如 v3.8.10-simplified.2
  --prerelease      将 GitHub Release 标记为预发布
  -h, --help        显示帮助

环境变量:
  CC                C 编译器，默认 gcc
  PYTHON            Python 可执行文件，默认 python
  ELECTRON_MIRROR   Electron 下载镜像（首次安装 Electron 时使用）
  ELECTRON_BUILDER_BINARIES_MIRROR  electron-builder 工具下载镜像
`);
    process.exit(0);
}

const fail = message => {
    console.error("错误: " + message);
    process.exit(1);
};
const quoteForCmd = value => /\s/.test(value) ? `"${value}"` : value;
const run = (command, commandArgs, options = {}) => {
    const printable = [command, ...commandArgs].join(" ");
    console.log("> " + printable);
    if (dryRun) return {status: 0, stdout: ""};
    const result = spawnSync(command, commandArgs, {stdio: "inherit", cwd: root, ...options});
    if (result.error) throw result.error;
    if (0 !== result.status) throw new Error(`命令失败 (${result.status}): ${printable}`);
    return result;
};
const runCapture = (command, commandArgs, options = {}) => {
    if (dryRun) return {status: 0, stdout: ""};
    return spawnSync(command, commandArgs, {encoding: "utf8", cwd: root, ...options});
};
const runPnpm = (pnpmArgs, options = {}) => {
    const execPath = process.env.npm_execpath;
    if (execPath) {
        // pnpm 的原生启动器是 .exe，JS 入口才需要用当前 Node 执行。
        if (/\.(?:mjs|cjs|js)$/i.test(execPath)) {
            return run(process.execPath, [execPath, ...pnpmArgs], options);
        }
        if (/\.cmd$/i.test(execPath)) {
            const commandLine = [execPath, ...pnpmArgs].map(quoteForCmd).join(" ");
            return run(process.env.ComSpec || "cmd.exe", ["/d", "/s", "/c", commandLine], options);
        }
        return run(execPath, pnpmArgs, options);
    }
    if ("win32" === process.platform) {
        const commandLine = ["pnpm", ...pnpmArgs.map(quoteForCmd)].join(" ");
        return run(process.env.ComSpec || "cmd.exe", ["/d", "/s", "/c", commandLine], options);
    }
    return run("pnpm", pnpmArgs, options);
};
const formatSize = bytes => {
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KiB";
    return (bytes / 1024 / 1024).toFixed(1) + " MiB";
};

if ("win32" !== process.platform) {
    fail("当前脚本只支持在 Windows 上生成 Windows 安装包");
}

const packageJson = JSON.parse(readFileSync(path.join(appDir, "package.json"), "utf8"));
const version = packageJson.version;
const nodeVersion = process.versions.node.split(".").map(Number);
if (24 !== nodeVersion[0] || 22 > nodeVersion[1]) {
    console.warn(`警告: 当前 Node ${process.versions.node}，建议 24.22+；较低版本的测试模块 mock 有缺陷`);
}
if (!existsSync(path.join(kernelDir, "go.mod"))) fail("找不到 kernel/go.mod，请在仓库根目录运行");
if (publish && !tag) fail("--publish 需要同时指定 --tag，例如 --tag v" + version + "-simplified.1");

const python = process.env.PYTHON || "python";
const cc = process.env.CC || "gcc";
const electronExe = path.join(appDir, "node_modules", "electron", "dist", "electron.exe");
const buildDir = path.join(appDir, "build");
const kernelOutput = path.join(appDir, "kernel", "SiYuan-Kernel.exe");

try {
    console.log(`SiYuan Simplified ${version} 本地 release 构建`);
    console.log(`仓库: ${root}`);

    if (!skipInstall) {
        runPnpm(["install", "--frozen-lockfile"], {cwd: appDir});
    }
    if (!existsSync(electronExe)) {
        const installScript = path.join(appDir, "node_modules", "electron", "install.js");
        if (!existsSync(installScript)) fail("缺少 Electron 依赖，请先运行 pnpm install");
        const env = {...process.env};
        if (process.env.ELECTRON_MIRROR) env.ELECTRON_MIRROR = process.env.ELECTRON_MIRROR;
        run(process.execPath, [installScript], {cwd: appDir, env});
    }

    // 每次都重建生成目录，避免旧产物混入安装包。
    for (const target of [buildDir, path.join(appDir, "stage", "build"), path.join(appDir, "electron-dist")]) {
        console.log("清理 " + path.relative(root, target));
        if (!dryRun) rmSync(target, {recursive: true, force: true});
    }
    if (!dryRun) mkdirSync(path.dirname(kernelOutput), {recursive: true});

    run(process.execPath, [path.join(appDir, "scripts", "clean-stage.mjs")], {cwd: appDir});
    for (const script of ["build:app", "build:mobile", "build:desktop", "build:export", "build:markdown", "build:electron"]) {
        runPnpm(["run", script], {cwd: appDir});
    }

    run(python, ["scripts/prepare-ocr.py", "--runtime", "windows-amd64"], {cwd: root});

    run("go", ["build", "-tags", "fts5 sqlcipher", "-ldflags", "-s -w", "-o", kernelOutput, "."], {
        cwd: kernelDir,
        env: {...process.env, CGO_ENABLED: "1", CC: cc, GOOS: "windows", GOARCH: "amd64"},
    });

    const builderArgs = ["exec", "electron-builder", "--config", "electron-builder.yml", "--publish=never"];
    const electronDist = path.join(appDir, "node_modules", "electron", "dist");
    if (existsSync(electronDist)) builderArgs.push("--config.electronDist=" + electronDist);
    runPnpm(builderArgs, {
        cwd: appDir,
        env: {...process.env, CSC_IDENTITY_AUTO_DISCOVERY: "false"},
    });

    const installer = readdirSync(buildDir).find(name =>
        /^siyuan-simplified-.+\.exe$/.test(name) && !name.includes("__uninstaller"));
    if (!installer) fail("构建结束但未找到安装包: " + buildDir);
    const installerPath = path.join(buildDir, installer);
    const hash = createHash("sha256").update(readFileSync(installerPath)).digest("hex");
    const checksumPath = installerPath + ".sha256";
    if (!dryRun) writeFileSync(checksumPath, `${hash}  ${installer}\n`, "utf8");

    const size = statSync(installerPath).size;
    console.log("");
    console.log("安装包: " + installerPath);
    console.log("大小: " + formatSize(size));
    console.log("SHA256: " + hash);
    console.log("校验文件: " + checksumPath);

    if (publish) {
        if (!dryRun) {
            const ghVersion = runCapture("gh", ["--version"]);
            if (ghVersion.error) fail("未找到 gh，请安装 GitHub CLI 并运行 gh auth login，或去掉 --publish");
            if (0 !== ghVersion.status) fail("gh 不可用，无法发布 GitHub Release");
            const auth = runCapture("gh", ["auth", "status"]);
            if (0 !== auth.status) fail("gh 未登录，请先运行 gh auth login");
        }
        const notesPath = path.join(buildDir, "release-notes.md");
        const commit = runCapture("git", ["rev-parse", "--short", "HEAD"]).stdout?.trim() || "unknown";
        if (!dryRun) {
            writeFileSync(notesPath, `# SiYuan Simplified ${tag}\n\n` +
                `- 版本: ${version}\n- 提交: ${commit}\n- 安装包: ${installer}\n- SHA256: ${hash}\n`, "utf8");
        }
        const releaseArgs = ["release", "create", tag, installerPath, checksumPath,
            "--title", `SiYuan Simplified ${tag}`, "--notes-file", notesPath];
        if (prerelease || tag.includes("-")) releaseArgs.push("--prerelease");
        const existing = dryRun ? {status: 1} : runCapture("gh", ["release", "view", tag]);
        if (0 === existing.status) {
            run("gh", ["release", "upload", tag, installerPath, checksumPath, "--clobber"]);
        } else {
            run("gh", releaseArgs);
        }
        console.log("已发布 GitHub Release: " + tag);
    } else {
        console.log("");
        console.log("未发布到 GitHub。需要本地发布时先安装并登录 gh，然后运行:");
        console.log(`  pnpm release:local -- --publish --tag v${version}-simplified.N`);
    }
} catch (error) {
    fail(error.message);
}
