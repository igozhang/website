#!/usr/bin/env node
/**
 * 将 igo-geo.ipgeo 复制到 internal/ipgeo/igo-geo-embed.ipgeo（go:embed 用）。
 * 用法：node scripts/gen-igo-geo-embed.mjs <源路径/igo-geo.ipgeo>
 */
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(__dirname, "..");
const src = process.argv[2];
const out = path.join(root, "internal/ipgeo/igo-geo-embed.ipgeo");

if (!src) {
  console.error("usage: node scripts/gen-igo-geo-embed.mjs <source.ipgeo>");
  process.exit(1);
}

const abs = path.resolve(src);
const buf = fs.readFileSync(abs);
if (buf.length < 64 || buf.slice(0, 5).toString() !== "IPGEO" || buf[5] !== 2) {
  console.error("invalid source: bad magic or version");
  process.exit(1);
}

fs.copyFileSync(abs, out);
console.log(`wrote ${out} (${buf.length} bytes)`);
