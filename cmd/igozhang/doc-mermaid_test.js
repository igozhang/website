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
if (!isMermaidBlock || !collect) {
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

console.log("OK: doc-mermaid helpers");
