#!/usr/bin/env node
// 清理构建产物。默认保留依赖和源码，只删除可再生成的产物。
import {existsSync, readdirSync, rmSync, statSync} from "node:fs";
import path from "node:path";
import {fileURLToPath} from "node:url";

const root = path.resolve(fileURLToPath(new URL("..", import.meta.url)));
const args = process.argv.slice(2);
const all = args.includes("--all");
const dryRun = args.includes("--dry-run");
const help = args.includes("--help") || args.includes("-h");

if (help) {
    console.log(`用法: pnpm clean:build [选项]

默认清理:
  app/build          安装包与打包产物
  app/stage/build    前端生产构建
  app/electron-dist  Electron 主进程编译产物
  app/kernel         内核可执行文件
  app/.eslintcache
  .cache             OCR 等下载缓存

选项:
  --all              额外删除 app/node_modules、app/stage/ocr 和 .pnpm-store
  --dry-run          只列出将删除的路径和大小
  -h, --help         显示帮助
`);
    process.exit(0);
}

const targets = [
    "app/build",
    "app/stage/build",
    "app/electron-dist",
    "app/kernel",
    "app/.eslintcache",
    ".cache",
];
if (all) {
    targets.push("app/node_modules", "app/stage/ocr", ".pnpm-store");
}

const sizeOf = target => {
    const info = statSync(target);
    if (!info.isDirectory()) return info.size;
    let total = 0;
    for (const entry of readdirSync(target, {withFileTypes: true})) {
        const child = path.join(target, entry.name);
        try {
            total += entry.isSymbolicLink() ? 0 : sizeOf(child);
        } catch {
            // 并发写入或权限问题不影响清理。
        }
    }
    return total;
};
const formatSize = bytes => {
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KiB";
    if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + " MiB";
    return (bytes / 1024 / 1024 / 1024).toFixed(2) + " GiB";
};

let freed = 0;
let removed = 0;
for (const relative of targets) {
    const target = path.resolve(root, relative);
    if (!target.startsWith(root + path.sep)) {
        console.error("跳过仓库外路径: " + target);
        continue;
    }
    if (!existsSync(target)) continue;
    let size = 0;
    try {
        size = sizeOf(target);
    } catch {
        // 大小统计失败时仍然清理。
    }
    console.log((dryRun ? "将删除 " : "删除 ") + relative + " (" + formatSize(size) + ")");
    if (!dryRun) rmSync(target, {recursive: true, force: true});
    freed += size;
    removed++;
}
console.log(`${dryRun ? "预计释放" : "已释放"} ${formatSize(freed)}，共 ${removed} 项`);
if (all) console.log("--all 已删除依赖，下次构建需要重新执行 pnpm install");
