"use strict";

const fs = require("fs");
const path = require("path");
const vm = require("vm");

const src = fs.readFileSync(path.join(__dirname, "doc-code-chrome.js"), "utf8");
const sandbox = {
  globalThis: {},
  window: undefined,
  document: { addEventListener: function () {} },
};
vm.runInNewContext(src, sandbox);
const highlightDocShell = sandbox.globalThis.highlightDocShell;
if (!highlightDocShell) {
  console.error("highlightDocShell not exported");
  process.exit(1);
}

const text = `cat > /igo/zabbix-phonecall.sh << 'SCRIPT_EOF'
#!/bin/bash
TO="\${1:-13088888888}"
PAYLOAD="{\\"called_number\\":\\"\${TO}\\"}"
wget -q --timeout=10 --header="Content-Type: application/json" --post-data="$PAYLOAD" -O /dev/null http://10.10.10.177:8080/call/tts
exit $?
SCRIPT_EOF`;

const html = highlightDocShell(text);

function decodeHtml(s) {
  return s
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&amp;/g, "&");
}

const visible = decodeHtml(html.replace(/<\/?span[^>]*>/g, ""));

if (visible !== text) {
  console.error("MISMATCH");
  console.error("expected first line:", text.split("\n")[0]);
  console.error("got first line:     ", visible.split("\n")[0]);
  process.exit(1);
}

if (html.includes("&gt;") || html.includes("&lt;")) {
  // entities in HTML source are OK if they decode back correctly
  if (visible !== text) {
    console.error("broken entities in output");
    process.exit(1);
  }
}

console.log("OK: visible text matches original");

// Comment line containing a quoted string (squid htpasswd password) must not leak H0/H1 placeholders.
const htpasswdLine =
  "# sudo htpasswd -bc /etc/squid/passwd igoproxy 'Igoproxy@54'";
const htHtml = highlightDocShell(htpasswdLine);
const htVisible = decodeHtml(htHtml.replace(/<\/?span[^>]*>/g, ""));
if (htVisible !== htpasswdLine) {
  console.error("HTPASSWD MISMATCH");
  console.error("expected:", htpasswdLine);
  console.error("got:     ", htVisible);
  process.exit(1);
}
console.log("OK: comment line with quoted password preserved");

// Go highlighter must not leak hN placeholders (type rule used to rematch H0/H5 inside markers).
const highlightDocCode = sandbox.globalThis.highlightDocCode;
if (!highlightDocCode) {
  console.error("highlightDocCode not exported");
  process.exit(1);
}
const goSrc = `// 时间复杂度：平均 O(nlogn)，最坏 O(n^2)
// 空间复杂度：平均 O(logn)
func QuickSort(arr []int) {
	if len(arr) == 0 {
		return
	}
}`;
const goHtml = highlightDocCode(goSrc, "go");
const goVisible = decodeHtml(goHtml.replace(/<\/?span[^>]*>/g, ""));
if (goVisible !== goSrc) {
  console.error("GO MISMATCH");
  console.error("expected:", goSrc);
  console.error("got:     ", goVisible);
  process.exit(1);
}
if (/\bh\d+\b/i.test(goVisible) || /H\d+/.test(goVisible)) {
  console.error("GO placeholder leak:", goVisible);
  process.exit(1);
}
if (!goHtml.includes("doc-hl-keyword") || !goHtml.includes("doc-hl-type")) {
  console.error("GO missing expected highlight classes");
  process.exit(1);
}
console.log("OK: Go QuickSort highlighting preserves source text");
