package main

import (
	_ "embed"
	"encoding/json"
	"html"
	"strings"
)

//go:embed site-foot-init.js
var siteFootInitJS []byte

const siteFootQuote = "一个男人，应该能够换尿布，策划战争，杀猪，开船，设计房子，写十四行诗，结算账户，砌墙，接脱臼的骨头，安慰濒死的人，服从命令，发布命令，携手合作，独立行动，解数学方程，分析新问题，铲粪，电脑编程，做出可口的饭，善打架，勇敢地死去。只有昆虫才专业化。"

const siteFootCSS = `
.site-foot{background:var(--panel);border-top:1px solid var(--line);padding:.85rem 0 1rem;margin-top:.25rem;flex-shrink:0}
.site-foot .shell{width:100%;max-width:var(--shell-max,52rem);margin:0 auto;padding-left:1.25rem;padding-right:1.25rem;box-sizing:border-box}
.site-foot-quote{font-size:.88rem;line-height:1.55;margin:.5rem 0 0;text-align:left}
.site-foot-row{display:flex;flex-wrap:wrap;gap:.5rem;align-items:center;justify-content:flex-start;margin-bottom:0}
.site-foot-games,.site-foot-dice-group{display:flex;flex-wrap:wrap;gap:.5rem;align-items:center}
.site-foot-games .phrase-next{flex:0 1 auto;white-space:nowrap}
.site-foot-dice-group{flex:1 1 11rem;min-width:0}
@media (max-width:480px){
.site-foot-games{gap:.35rem .4rem}
.site-foot-games .phrase-next{font-size:.78rem;padding:.24rem .5rem}
}
@media (max-width:400px){
.site-foot-games{flex:1 1 100%}
.site-foot-dice-group{flex:1 1 100%}
}
.site-foot-row .phrase-next{font:inherit;font-size:.8rem;padding:.28rem .65rem;border-radius:6px;border:1px solid var(--line);background:var(--panel);color:var(--accent);cursor:pointer;text-decoration:none;display:inline-block}
.site-foot-row button.phrase-next{font-family:inherit}
a.phrase-next:visited{color:var(--accent)}
.foot-dice-last{flex:0 1 auto;min-width:6.5rem;font-size:.82rem;color:var(--muted);margin:0;padding:.28rem .5rem;border-radius:6px;border:1px dashed var(--line);background:rgba(255,255,255,.25)}
.foot-dice-last[hidden]{display:none!important}
.foot-dice-last strong{font-size:1.05rem;color:var(--accent);font-weight:700;font-variant-numeric:tabular-nums;margin-left:.15rem}
.foot-dice-overlay{position:fixed;inset:0;z-index:10001;display:none;align-items:center;justify-content:center;background:rgba(36,48,36,.38);pointer-events:none}
.foot-dice-overlay.is-open{display:flex}
.foot-dice-overlay-inner{perspective:520px;display:flex;align-items:center;justify-content:center}
.foot-dice-cube{width:4.25rem;height:4.25rem;position:relative;transform-style:preserve-3d;transform:rotateX(-18deg) rotateY(24deg);will-change:transform}
.foot-dice-cube.foot-dice-rolling{animation:footDiceRoll3d 1.05s ease-in-out infinite}
.foot-dice-cube.foot-dice-settled{transition:transform .55s cubic-bezier(.2,.85,.25,1)}
@keyframes footDiceRoll3d{
  0%{transform:rotateX(0deg) rotateY(0deg) rotateZ(0deg)}
  25%{transform:rotateX(144deg) rotateY(108deg) rotateZ(36deg)}
  50%{transform:rotateX(252deg) rotateY(216deg) rotateZ(72deg)}
  75%{transform:rotateX(396deg) rotateY(324deg) rotateZ(108deg)}
  100%{transform:rotateX(540deg) rotateY(432deg) rotateZ(144deg)}
}
.foot-dice-face{position:absolute;inset:0;border-radius:7px;background:linear-gradient(145deg,#fafcf8,#d8e6d8);border:1px solid #9ab09a;
  box-shadow:inset 0 1px 0 rgba(255,255,255,.75);display:grid;grid-template:repeat(3,1fr)/repeat(3,1fr);padding:3px;backface-visibility:hidden}
.foot-dice-face i{display:block;width:.52rem;height:.52rem;border-radius:50%;background:#2a3d2a;align-self:center;justify-self:center;box-shadow:inset 0 1px 1px rgba(0,0,0,.25)}
.foot-dice-face.f1{transform:rotateY(0deg) translateZ(2.125rem)}
.foot-dice-face.f2{transform:rotateY(90deg) translateZ(2.125rem)}
.foot-dice-face.f3{transform:rotateX(90deg) translateZ(2.125rem)}
.foot-dice-face.f4{transform:rotateX(-90deg) translateZ(2.125rem)}
.foot-dice-face.f5{transform:rotateY(-90deg) translateZ(2.125rem)}
.foot-dice-face.f6{transform:rotateY(180deg) translateZ(2.125rem)}
.foot-dice-face.f1 i:only-child{grid-area:2/2}
.foot-dice-face.f2 i:nth-child(1){grid-area:1/1}.foot-dice-face.f2 i:nth-child(2){grid-area:3/3}
.foot-dice-face.f3 i:nth-child(1){grid-area:1/1}.foot-dice-face.f3 i:nth-child(2){grid-area:2/2}.foot-dice-face.f3 i:nth-child(3){grid-area:3/3}
.foot-dice-face.f4 i:nth-child(1){grid-area:1/1}.foot-dice-face.f4 i:nth-child(2){grid-area:1/3}.foot-dice-face.f4 i:nth-child(3){grid-area:3/1}.foot-dice-face.f4 i:nth-child(4){grid-area:3/3}
.foot-dice-face.f5 i:nth-child(1){grid-area:1/1}.foot-dice-face.f5 i:nth-child(2){grid-area:1/3}.foot-dice-face.f5 i:nth-child(3){grid-area:2/2}.foot-dice-face.f5 i:nth-child(4){grid-area:3/1}.foot-dice-face.f5 i:nth-child(5){grid-area:3/3}
.foot-dice-face.f6 i:nth-child(1){grid-area:1/1}.foot-dice-face.f6 i:nth-child(2){grid-area:1/2}.foot-dice-face.f6 i:nth-child(3){grid-area:1/3}
.foot-dice-face.f6 i:nth-child(4){grid-area:3/1}.foot-dice-face.f6 i:nth-child(5){grid-area:3/2}.foot-dice-face.f6 i:nth-child(6){grid-area:3/3}
.site-foot-links{container-type:inline-size;display:grid;grid-template-columns:minmax(0,1fr) auto;align-items:baseline;gap:.15rem 1.1rem;margin-top:.45rem;padding-top:.55rem;border-top:1px solid var(--line);font-size:.84rem;line-height:1.5}
.site-foot-links-nav{display:flex;flex-wrap:wrap;align-items:baseline;justify-content:flex-start;gap:.15rem .05rem;min-width:0}
.site-foot-links-nav a{color:var(--accent);text-decoration:none;padding:.12rem .35rem;border-radius:4px}
.site-foot-links-nav a:hover{text-decoration:underline;background:rgba(61,107,74,.08)}
.site-foot-links-nav .sep{color:var(--muted);user-select:none;padding:0 .12rem;font-size:.78rem;opacity:.7}
.site-foot-visitor{justify-self:end;max-width:100%;font-size:.76rem;color:var(--muted);white-space:nowrap;font-variant-numeric:tabular-nums;letter-spacing:.01em;padding:.1rem 0;opacity:.9}
@container (max-width:30rem){
.site-foot-links{grid-template-columns:1fr;gap:.12rem}
.site-foot-visitor{justify-self:start;white-space:normal;word-break:break-all}
}
@media (max-width:480px){
.site-foot-links{grid-template-columns:1fr;gap:.12rem}
.site-foot-visitor{justify-self:start;white-space:normal;word-break:break-all}
}
.phrase-en{font-style:italic;color:var(--text)}
.site-foot-quote .foot-quote-hit{font:inherit;font-style:inherit;font-size:inherit;line-height:inherit;color:inherit;background:transparent;border:0;padding:0;margin:0;text-decoration:none;cursor:inherit;-webkit-tap-highlight-color:transparent}
.site-foot-quote .foot-quote-hit:hover,.site-foot-quote .foot-quote-hit:active,.site-foot-quote .foot-quote-hit:focus{color:inherit;background:transparent;text-decoration:none;outline:none;box-shadow:none}
@media (prefers-reduced-motion:reduce){
  .foot-dice-cube.foot-dice-rolling{animation:none}
  .foot-dice-cube.foot-dice-settled{transition-duration:.01ms}
}
`

const siteFootQuoteIgoHit = "专业化"

func siteFootQuoteHTML() string {
	igoIdx := strings.LastIndex(siteFootQuote, siteFootQuoteIgoHit)
	if igoIdx < 0 {
		return html.EscapeString(siteFootQuote)
	}
	var b strings.Builder
	b.WriteString(html.EscapeString(siteFootQuote[:igoIdx]))
	b.WriteString(`<span class="foot-quote-hit" id="footIgoHit" aria-hidden="true">`)
	b.WriteString(siteFootQuoteIgoHit)
	b.WriteString(`</span>`)
	b.WriteString(html.EscapeString(siteFootQuote[igoIdx+len(siteFootQuoteIgoHit):]))
	return b.String()
}

func siteFootBlock(pub, gomokuHref, xiangqiHref, texasHref, zhijianHref, visitorLabel string) string {
	escPub := html.EscapeString(pub)
	escGomoku := html.EscapeString(gomokuHref)
	escXiangqi := html.EscapeString(xiangqiHref)
	escTexas := html.EscapeString(texasHref)
	escZhijian := html.EscapeString(zhijianHref)
	quote := siteFootQuoteHTML()
	baseJS, _ := json.Marshal(pub)
	var b []byte
	b = append(b, `<footer class="site-foot" data-base="`...)
	b = append(b, escPub...)
	b = append(b, `">
<div class="shell">
<div class="site-foot-row">
<div class="site-foot-games">
<a class="phrase-next" id="footGomoku" href="`...)
	b = append(b, escGomoku...)
	b = append(b, `" title="五子棋人机对战">五子棋</a>
<a class="phrase-next" id="footXiangqi" href="`...)
	b = append(b, escXiangqi...)
	b = append(b, `" title="象棋人机对战">象棋</a>
<a class="phrase-next" id="footTexas" href="`...)
	b = append(b, escTexas...)
	b = append(b, `" title="德州扑克人机对战">德扑</a>
<a class="phrase-next" id="footZhijian" href="`...)
	b = append(b, escZhijian...)
	b = append(b, `" title="知键 · 英文字母打字练习">知键</a>
</div>
<div class="site-foot-dice-group">
<button type="button" class="phrase-next" id="footDice" title="掷骰子">骰子</button>
<span class="foot-dice-last" id="footDiceLastWrap" hidden>上次点数<strong id="footDiceLastVal">—</strong></span>
</div>
</div>
<p class="site-foot-quote phrase-en">`...)
	b = append(b, quote...)
	b = append(b, `</p>
<div class="site-foot-links">
<nav class="site-foot-links-nav" id="footLinks" aria-label="站点链接">
<a href="https://igozhang.cn" target="_blank" rel="noopener noreferrer">igozhang.cn</a>
</nav>`...)
	if visitorLabel != "" {
		b = append(b, `<span class="site-foot-visitor" title="来访 IP">`...)
		b = append(b, visitorLabel...)
		b = append(b, `</span>`...)
	}
	b = append(b, `
</div>
</div>
</footer>
<div class="foot-dice-overlay" id="footDiceOverlay" aria-hidden="true">
<div class="foot-dice-overlay-inner">
<div class="foot-dice-cube" id="footDiceCube">
<div class="foot-dice-face f1" aria-hidden="true"><i></i></div>
<div class="foot-dice-face f2" aria-hidden="true"><i></i><i></i></div>
<div class="foot-dice-face f3" aria-hidden="true"><i></i><i></i><i></i></div>
<div class="foot-dice-face f4" aria-hidden="true"><i></i><i></i><i></i><i></i></div>
<div class="foot-dice-face f5" aria-hidden="true"><i></i><i></i><i></i><i></i><i></i></div>
<div class="foot-dice-face f6" aria-hidden="true"><i></i><i></i><i></i><i></i><i></i><i></i></div>
</div>
</div>
</div>
`...)
	b = append(b, opsDiceInjectScript()...)
	b = append(b, `<script>`...)
	b = append(b, siteFootInitJS...)
	b = append(b, `</script>
<script>window.__fsSiteFootInit&&window.__fsSiteFootInit(`...)
	b = append(b, baseJS...)
	b = append(b, `);</script>`...)
	return string(b)
}
