package main

import _ "embed"

//go:embed site-ui-guard.js
var siteUIGuardJS []byte

// siteUIGuardCSS 全站 UI 防护：非编辑区不显示文本插入符（caret），静态文字仍可正常选中复制。
//
// 插入符抑制主要靠 siteUIGuardScript（focusin 收回误焦点）；caret-color 仅作视觉兜底。
const siteUIGuardCSS = `
body,:where(h1,h2,h3,h4,h5,h6,p,span,li,div,blockquote,pre,code,small,strong,em,td,th,dt,dd,
.site-head,.site-foot,.phrase-line,.site-foot-quote,.msg,.toast-wrap,
main.shell,section,.zhijian-page,.gomoku-page,.xiangqi-page,.texas-page,.zhijian-letter,.zhijian-hint,.zj-letter-top,.zj-letter-foot,.zj-material,.zj-mat-char,
.mp3-section,.mp3-title,.mp3-meta):not(input):not(textarea):not(select):not([contenteditable="true"]){
  caret-color:transparent
}
input:not([type=button]):not([type=submit]):not([type=reset]):not([type=checkbox]):not([type=radio]):not([type=file]):not([type=hidden]):not([type=image]),
textarea,.stash-textarea,select,[contenteditable="true"]{
  caret-color:auto
}
button,.phrase-next,.zhijian-btn,a.phrase-next,a[href],.site-foot-row .phrase-next,.site-foot-row button.phrase-next,
.site-foot-links a,.zhijian-key,.zhijian-infinite-btn,label,summary,[role="button"]{
  caret-color:transparent
}
.fs-decor-layer,.zj-fuse,.zj-fuse-track,.zj-fireworks,.foot-dice-overlay{
  overflow:hidden
}
`

// siteUIGuardScript 返回全站 caret 防护脚本（独立 <script>，勿嵌套进其它脚本）。
func siteUIGuardScript() string {
	return "<script>" + string(siteUIGuardJS) + "</script>"
}
