package main

import _ "embed"

//go:embed doc-code-chrome.js
var docCodeChromeJS []byte

// docCodeBlockCSS styles fenced code on article pages (VS Code Dark+ shell palette + light lang tokens).
func docCodeBlockCSS() string {
	return `
.doc-code-wrap{position:relative;margin:.75rem 0;
  --out-bg:#0d1117;--out-fg:#c9d1d9;--out-border:#30363d;
  --hl-comment:#6a9955;--hl-string:#ce9178;--hl-keyword:#569cd6;--hl-cmd:#dcdcaa;
  --hl-flag:#9cdcfe;--hl-var:#4ec9b0;--hl-num:#b5cea8;--hl-key:#79c0ff;--hl-bool:#569cd6;
  --hl-type:#4ec9b0;--hl-func:#dcdcaa;--hl-tag:#7ee787;--hl-attr:#9cdcfe;--hl-op:#8b949e;
  --hl-time:#569cd6;--hl-path:#79c0ff;--hl-ip:#a5d6ff;--hl-port:#79c0ff;
  --hl-unit:#d2a8ff;--hl-pid:#e6edf3;--hl-user:#7ee787;
  --hl-ok:#3fb950;--hl-warn:#d29922;--hl-err:#ff7b72;--hl-dim:#8b949e;
}
.doc-code-wrap:hover .doc-copy-btn,.doc-code-wrap:focus-within .doc-copy-btn{opacity:1;pointer-events:auto}
.doc-copy-btn{position:absolute;top:.4rem;right:.4rem;z-index:2;opacity:0;pointer-events:none;transition:opacity .15s ease;
  font:inherit;font-size:.72rem;line-height:1;padding:.28rem .5rem;border:1px solid rgba(255,255,255,.18);border-radius:5px;
  background:rgba(0,0,0,.55);color:#e8e8e8;cursor:pointer;backdrop-filter:blur(4px)}
.doc-copy-btn:hover{background:rgba(0,0,0,.72);border-color:rgba(255,255,255,.32)}
.doc-md pre,.doc-md .doc-code-wrap pre{
  margin:0;background:var(--out-bg);color:var(--out-fg);border:1px solid var(--out-border);border-radius:6px;
  padding:.65rem .75rem;padding-top:2rem;overflow:auto;max-width:100%;
  font-family:ui-monospace,Consolas,"Cascadia Mono",monospace;font-size:.82rem;line-height:1.5;
  white-space:pre;word-break:normal;overflow-wrap:normal;tab-size:4}
.doc-md pre code{background:transparent;color:inherit;font-size:inherit;padding:0;white-space:inherit}
.doc-md :not(pre)>code{
  background:var(--doc-surface);color:var(--text);padding:.12rem .35rem;border-radius:4px;
  white-space:pre-wrap;word-break:break-word;font-size:.88em}
.doc-code-wrap .doc-hl-comment{color:var(--hl-comment);font-style:italic}
.doc-code-wrap .doc-hl-string{color:var(--hl-string)}
.doc-code-wrap .doc-hl-keyword{color:var(--hl-keyword);font-weight:600}
.doc-code-wrap .doc-hl-cmd{color:var(--hl-cmd);font-weight:600}
.doc-code-wrap .doc-hl-flag{color:var(--hl-flag)}
.doc-code-wrap .doc-hl-var{color:var(--hl-var)}
.doc-code-wrap .doc-hl-num{color:var(--hl-num)}
.doc-code-wrap .doc-hl-key{color:var(--hl-key)}
.doc-code-wrap .doc-hl-bool{color:var(--hl-bool);font-weight:600}
.doc-code-wrap .doc-hl-type{color:var(--hl-type)}
.doc-code-wrap .doc-hl-func{color:var(--hl-func);font-weight:600}
.doc-code-wrap .doc-hl-tag{color:var(--hl-tag)}
.doc-code-wrap .doc-hl-attr{color:var(--hl-attr)}
.doc-code-wrap .doc-hl-op{color:var(--hl-op)}
.doc-code-wrap .doc-hl-time{color:var(--hl-time)}
.doc-code-wrap .doc-hl-path{color:var(--hl-path)}
.doc-code-wrap .doc-hl-ip{color:var(--hl-ip)}
.doc-code-wrap .doc-hl-port{color:var(--hl-port)}
.doc-code-wrap .doc-hl-unit{color:var(--hl-unit)}
.doc-code-wrap .doc-hl-pid{color:var(--hl-pid);opacity:.92}
.doc-code-wrap .doc-hl-user{color:var(--hl-user)}
.doc-code-wrap .doc-hl-ok{color:var(--hl-ok);font-weight:600}
.doc-code-wrap .doc-hl-warn{color:var(--hl-warn);font-weight:600}
.doc-code-wrap .doc-hl-err{color:var(--hl-err);font-weight:600}
.doc-code-wrap .doc-hl-dim{color:var(--hl-dim)}
` + docMermaidCSS()
}

func docCodeChromeScript() string {
	return "<script>" + string(docCodeChromeJS) + "</script>"
}

// docCodeCopyScript is an alias for article pages (highlight + copy).
func docCodeCopyScript() string {
	return docCodeChromeScript()
}
