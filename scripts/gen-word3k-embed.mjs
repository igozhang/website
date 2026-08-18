#!/usr/bin/env node
/**
 * 从 assets/eng/word3k-full.md 随机抽取 50 条有效词条，写入 cmd/igozhang/word3k-embed.md（go:embed 用）。
 * 用法：node scripts/gen-word3k-embed.mjs
 */
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(__dirname, "..");
const src = path.join(root, "assets/eng/word3k-full.md");
const out = path.join(root, "cmd/igozhang/word3k-embed.md");
const PIPE = "\uFF5C";
const COUNT = 50;
const SEED = 20260523;

function parseHead(s) {
  return /^(\S+)\s+(\/.+?\/)\s+(.+)$/.test(s);
}

function hasPracticeLetters(word, def, ex) {
  let n = 0;
  for (const part of [word, def, ex]) {
    for (const c of part) if (/[a-zA-Z]/.test(c)) n++;
  }
  return n > 0;
}

const text = fs.readFileSync(src, "utf8");
const valid = [];
for (const line of text.split(/\r?\n/)) {
  const raw = line.trim();
  if (!raw) continue;
  const parts = raw.split(PIPE);
  if (parts.length !== 4 || parts.some((p) => !p.trim())) continue;
  const head = parts[0].trim();
  if (!parseHead(head)) continue;
  const word = head.match(/^(\S+)/)[1];
  if (!hasPracticeLetters(word, parts[1], parts[2])) continue;
  valid.push(raw);
}

let s = SEED;
function rnd() {
  s = (s * 1103515245 + 12345) & 0x7fffffff;
  return s / 0x7fffffff;
}

const order = valid.map((_, i) => i);
for (let i = order.length - 1; i > 0; i--) {
  const j = Math.floor(rnd() * (i + 1));
  [order[i], order[j]] = [order[j], order[i]];
}
const picked = order
  .slice(0, COUNT)
  .sort((a, b) => a - b)
  .map((i) => valid[i]);

fs.writeFileSync(out, picked.join("\n") + "\n");
console.log(`wrote ${picked.length} lines (${fs.statSync(out).size} bytes) -> ${out}`);
