"use strict";

const fs = require("fs");
const path = require("path");
const vm = require("vm");

const src = fs.readFileSync(path.join(__dirname, "doc-mermaid.js"), "utf8");
const sandbox = {
  globalThis: {},
  window: undefined,
  document: { addEventListener: function () {} },
};
vm.runInNewContext(src, sandbox);

const isMermaidBlock = sandbox.globalThis.__igoIsMermaidBlock;
const collect = sandbox.globalThis.__igoCollectMermaidBlocks;
const sanitizeSvg = sandbox.globalThis.__igoSanitizeSvgForExport;
const sizeFromString = sandbox.globalThis.__igoSvgExportSizeFromString;
const clampScale = sandbox.globalThis.__igoClampExportScale;
if (!isMermaidBlock || !collect || !sanitizeSvg || !sizeFromString || !clampScale) {
  console.error("helpers not exported");
  process.exit(1);
}

function fakePre(codeClass, text) {
  const code = {
    className: codeClass || "",
    textContent: text || "",
  };
  const pre = {
    className: "",
    querySelector: function (sel) {
      return sel === "code" ? code : null;
    },
    closest: function () {
      return null;
    },
    textContent: text || "",
  };
  return pre;
}

if (!isMermaidBlock(fakePre("language-mermaid", "flowchart TB\nA-->B"))) {
  console.error("should detect language-mermaid");
  process.exit(1);
}
if (isMermaidBlock(fakePre("language-bash", "echo hi"))) {
  console.error("must not detect bash as mermaid");
  process.exit(1);
}

const root = {
  querySelectorAll: function () {
    return [fakePre("language-mermaid", "graph LR\nA-->B"), fakePre("language-go", "package main")];
  },
};
const blocks = collect(root);
if (blocks.length !== 1 || blocks[0].text.indexOf("graph LR") < 0) {
  console.error("collectMermaidBlocks failed", blocks);
  process.exit(1);
}

const dirty =
  '<svg viewBox="0 0 100 40" width="100%" xmlns="http://www.w3.org/2000/svg"><foreignObject><p>a<br>b</p></foreignObject></svg>';
const cleaned = sanitizeSvg(dirty);
if (!/<br\/>/.test(cleaned) || /<br>/.test(cleaned.replace(/<br\/>/g, ""))) {
  console.error("sanitizeSvgForExport should normalize <br>", cleaned);
  process.exit(1);
}
const sized = sizeFromString(dirty);
if (!sized || sized.w !== 100 || sized.h !== 40) {
  console.error("svgExportSizeFromString should read viewBox via regex", sized);
  process.exit(1);
}
const scaled = clampScale({ w: 5284, h: 1839 }, 2);
if (!(scaled < 2) || !(scaled > 1)) {
  console.error("clampExportScale should reduce scale for huge diagrams", scaled);
  process.exit(1);
}
const oversized = { w: 200000, h: 1000 };
const tinyScale = clampScale(oversized, 2);
const tinyEdge = Math.ceil(oversized.w * tinyScale);
if (!(tinyScale < 0.05) || tinyEdge > 8192) {
  console.error("clampExportScale must not exceed max edge via a scale floor", tinyScale, tinyEdge);
  process.exit(1);
}
if (sanitizeSvg("<browser>x</browser>") !== "<browser>x</browser>") {
  console.error("sanitize must not touch <browser>");
  process.exit(1);
}

console.log("OK: doc-mermaid helpers");
