package main

import _ "embed"

//go:embed ui/igo-auth.js
var igoAuthJS []byte

// igoAuthPanelHTML 三端统一的密钥认证入口（标题固定为「IGO密钥认证」）。
const igoAuthPanelHTML = `<div id="igoAuthSec" class="igo-auth" aria-label="IGO密钥认证">
<h2>IGO密钥认证</h2>
<div class="igo-auth-field">
<label for="igoKey">验证密钥（key）</label>
<input type="text" class="igo-auth-input" id="igoKey" autocomplete="off" spellcheck="false" maxlength="8" placeholder="Input your password ：">
</div>
<div class="igo-auth-actions">
<button type="button" class="igo-auth-btn" id="btnIgoAuth">验证</button>
<span class="igo-auth-expiry" id="igoKeyExpiry" aria-live="polite"></span>
</div>
<p class="igo-auth-msg" id="igoAuthMsg" aria-live="polite"></p>
</div>`

const igoAuthCSS = `
.igo-auth{margin:0 0 1rem}
.igo-auth h2{font-size:1.05rem;margin:0 0 .75rem;font-weight:600;text-align:center;color:var(--text,#2d3b2d);text-transform:none;letter-spacing:0}
.igo-auth-field{margin:.65rem 0}
.igo-auth-field label{display:block;font-size:.84rem;color:var(--muted,#5a6b5a);margin-bottom:.25rem}
.igo-auth-input{display:block;width:100%;max-width:100%;min-height:2.35rem;font:inherit;font-size:.9rem;padding:.45rem .55rem;border:1px solid var(--line,#b8c9b8);border-radius:6px;background:var(--input,var(--bg,#e8efe8));color:var(--text,#2d3b2d);box-sizing:border-box}
.igo-auth-actions{display:flex;flex-wrap:wrap;align-items:center;gap:.5rem;margin-top:.35rem}
.igo-auth-btn{font:inherit;font-size:.86rem;padding:.38rem .9rem;border-radius:6px;border:1px solid var(--line,#b8c9b8);background:var(--panel,#dde8dd);color:var(--accent,#3d6b4a);cursor:pointer}
.igo-auth-btn:hover{background:rgba(61,107,74,.1)}
.igo-auth-msg{font-size:.88rem;margin-top:.5rem;min-height:1.2em;color:var(--muted,#5a6b5a)}
.igo-auth-msg.is-ok{color:var(--ok,#2d5a3a)}
.igo-auth-msg.is-err{color:var(--err,#7a3a2a)}
.igo-auth-expiry{font-size:.82rem;color:var(--muted,#5a6b5a);line-height:1.4;flex:1 1 auto}
.igo-auth-expiry.is-warn{color:var(--err,#7a3a2a)}
.igo-authed-hint{font-size:.88rem;line-height:1.5;margin:0 0 .85rem;color:var(--ok,#2d5a3a)}
.hidden{display:none!important}
`

func igoAuthScriptTag() string {
	return "<script>" + string(igoAuthJS) + "</script>"
}
