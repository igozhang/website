#!/usr/bin/env node
/**
 * 从 assets/eng/phrases.md 取前 100 条有效引语，写入 cmd/igozhang/phrases-embed.md（go:embed 用）。
 * 用法：node scripts/gen-phrases-embed.mjs
 */
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(__dirname, "..");
const src = path.join(root, "assets/eng/phrases.md");
const out = path.join(root, "cmd/igozhang/phrases-embed.md");
const PIPE = "\uFF5C";
const COUNT = 100;

const text = fs.readFileSync(src, "utf8");
const lines = [];
for (const line of text.split(/\r?\n/)) {
  const raw = line.trim();
  if (!raw || raw.startsWith("#")) continue;
  const parts = raw.split(PIPE);
  if (parts.length !== 2 || !parts[0].trim() || !parts[1].trim()) continue;
  lines.push(`${parts[0].trim()}${PIPE}${parts[1].trim()}`);
  if (lines.length >= COUNT) break;
}

if (lines.length < COUNT) {
  console.error(`only ${lines.length} valid lines in ${src}, need ${COUNT}`);
  process.exit(1);
}

const body = [
  "# 内嵌引语精简副本（100 条）；完整库见 DATA_ROOT/assets/eng/phrases.md",
  "# 重新生成：node scripts/gen-phrases-embed.mjs",
  "",
  ...lines,
  "",
].join("\n");

fs.writeFileSync(out, body, "utf8");
console.log(`wrote ${out} (${lines.length} phrases)`);
