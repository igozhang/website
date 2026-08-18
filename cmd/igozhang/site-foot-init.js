window.__fsSiteFootInit=function(base){
  var gomokuEl=document.getElementById("footGomoku");
  var xiangqiEl=document.getElementById("footXiangqi");
  var texasEl=document.getElementById("footTexas");
  var zhijianEl=document.getElementById("footZhijian");
  var diceBtn=document.getElementById("footDice");
  var diceOverlay=document.getElementById("footDiceOverlay");
  var diceCube=document.getElementById("footDiceCube");
  var diceLastWrap=document.getElementById("footDiceLastWrap");
  var diceLastVal=document.getElementById("footDiceLastVal");
  var _diceOps=(typeof window!=="undefined"&&window.__fsDiceOps)||{};
  var DICE_ROLL_MS=_diceOps.rollMs||1150;
  var DICE_RESULT_MS=_diceOps.resultShowMs||2200;
  var diceRolling=false;
  var diceTimers=[];
  var diceFaceRot={
    1:"rotateX(-18deg) rotateY(24deg)",
    2:"rotateX(-18deg) rotateY(-66deg)",
    3:"rotateX(-108deg) rotateY(24deg)",
    4:"rotateX(72deg) rotateY(24deg)",
    5:"rotateX(-18deg) rotateY(114deg)",
    6:"rotateX(-18deg) rotateY(204deg)"
  };
  function absRoot(){
    if(typeof window.__fsAbsRoot==="function") return window.__fsAbsRoot(base);
    if(base!==undefined&&base!==null&&String(base).length) return String(base).replace(/\/$/,"");
    return "";
  }
  function footLooksLikeIgoKey(s){
    return /^[a-zA-Z]{8}$/.test(String(s||"").trim());
  }
  function footHaptic(){
    try{ if(typeof navigator.vibrate==="function") navigator.vibrate(14); }catch(e){}
  }
  function footClearTimers(){
    for(var i=0;i<diceTimers.length;i++) clearTimeout(diceTimers[i]);
    diceTimers=[];
  }
  function footResolveHref(href){
    if(!href) return "";
    if(/^https?:\/\//i.test(href)) return href;
    var root=absRoot();
    if(href.charAt(0)==="/"){
      var p=(root||"")+href;
      try{ return new URL(p, location.origin||undefined).href; }catch(e){ return (location.origin||"")+p; }
    }
    var pre=root?root+"/":"/";
    try{ return new URL(pre+href.replace(/^\//,""), location.origin||undefined).href; }catch(e){ return (location.origin||"")+pre+href; }
  }
  function renderFootLinks(links){
    var nav=document.getElementById("footLinks");
    if(!nav) return;
    nav.textContent="";
    if(!links||!links.length) return;
    links.forEach(function(l, idx){
      if(idx>0){
        var sep=document.createElement("span");
        sep.className="sep";
        sep.textContent="·";
        sep.setAttribute("aria-hidden","true");
        nav.appendChild(sep);
      }
      var a=document.createElement("a");
      a.href=footResolveHref(l.href);
      a.target="_blank";
      a.rel="noopener noreferrer";
      a.textContent=l.text||l.href;
      nav.appendChild(a);
    });
  }
  fetch(absRoot()+"/api/footer",{cache:"no-store"})
    .then(function(r){ return r.ok?r.json():null; })
    .then(function(cfg){
      if(!cfg) return;
      if(gomokuEl&&cfg.gomoku) gomokuEl.href=footResolveHref(cfg.gomoku);
      if(xiangqiEl&&cfg.xiangqi) xiangqiEl.href=footResolveHref(cfg.xiangqi);
      else if(xiangqiEl) xiangqiEl.href=footResolveHref("/xiangqi");
      if(texasEl&&cfg.texas) texasEl.href=footResolveHref(cfg.texas);
      else if(texasEl) texasEl.href=footResolveHref("/texas");
      if(zhijianEl) zhijianEl.href=footResolveHref("/zhijian");
      if(cfg.links&&cfg.links.length) renderFootLinks(cfg.links);
    })
    .catch(function(){});
  function footDiceHideOverlay(){
    if(diceOverlay){ diceOverlay.classList.remove("is-open"); diceOverlay.setAttribute("aria-hidden","true"); }
    if(diceCube){
      diceCube.classList.remove("foot-dice-rolling","foot-dice-settled");
      diceCube.style.transform="";
    }
  }
  function footDiceReset(){
    diceRolling=false;
    footDiceHideOverlay();
    if(diceBtn) diceBtn.disabled=false;
  }
  function footDiceShowResult(n){
    if(diceCube){
      diceCube.classList.remove("foot-dice-rolling");
      void diceCube.offsetWidth;
      diceCube.classList.add("foot-dice-settled");
      diceCube.style.transform=diceFaceRot[n]||diceFaceRot[1];
    }
    if(diceLastVal) diceLastVal.textContent=String(n);
    if(diceLastWrap) diceLastWrap.hidden=false;
    footHaptic();
    diceTimers.push(setTimeout(footDiceReset, DICE_RESULT_MS));
  }
  function footDiceRoll(){
    if(!diceBtn||!diceCube||!diceOverlay||diceRolling) return;
    footHaptic();
    footClearTimers();
    diceRolling=true;
    diceBtn.disabled=true;
    var finalN=1+Math.floor(Math.random()*6);
    var reduced=typeof window.matchMedia==="function"&&window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    diceOverlay.classList.add("is-open");
    diceOverlay.setAttribute("aria-hidden","false");
    diceCube.classList.remove("foot-dice-settled","foot-dice-rolling");
    diceCube.style.transform="";
    if(reduced){ footDiceShowResult(finalN); return; }
    requestAnimationFrame(function(){
      requestAnimationFrame(function(){ diceCube.classList.add("foot-dice-rolling"); });
    });
    diceTimers.push(setTimeout(function(){ footDiceShowResult(finalN); }, DICE_ROLL_MS));
  }
  if(diceBtn) diceBtn.addEventListener("click", footDiceRoll);
  var igoHit=document.getElementById("footIgoHit");
  var igoFetchPending=false;
  var igoKeyCache="";
  var igoKeyCacheAt=0;
  var IGO_KEY_CACHE_MS=5*60*1000;
  var FOOT_HIT_WINDOW_MS=2000;
  function footBindDoubleHit(el, onTrigger){
    if(!el||typeof onTrigger!=="function") return;
    var clicks=0, timer=null;
    el.addEventListener("click", function(ev){
      ev.preventDefault();
      clicks+=1;
      if(timer) clearTimeout(timer);
      if(clicks>=2){
        clicks=0;
        timer=null;
        onTrigger();
        return;
      }
      timer=setTimeout(function(){
        clicks=0;
        timer=null;
      }, FOOT_HIT_WINDOW_MS);
    });
  }
  function footCopyPlain(text){
    if(!text) return Promise.reject(new Error("empty"));
    var ta=document.createElement("textarea");
    ta.value=text;
    ta.setAttribute("readonly","");
    ta.style.position="fixed";
    ta.style.top="0";
    ta.style.left="0";
    ta.style.width="2em";
    ta.style.height="2em";
    ta.style.padding="0";
    ta.style.border="none";
    ta.style.outline="none";
    ta.style.boxShadow="none";
    ta.style.background="transparent";
    ta.style.opacity="0";
    ta.style.pointerEvents="none";
    document.body.appendChild(ta);
    ta.focus();
    ta.select();
    try{ ta.setSelectionRange(0, text.length); }catch(e){}
    var ok=false;
    try{ ok=document.execCommand("copy"); }catch(e){}
    document.body.removeChild(ta);
    if(ok) return Promise.resolve();
    if(navigator.clipboard&&navigator.clipboard.writeText){
      return navigator.clipboard.writeText(text);
    }
    return Promise.reject(new Error("copy failed"));
  }
  function footIgoCopyToast(text,isErr){
    try{
      if(window.__fs&&typeof window.__fs.showToast==="function") window.__fs.showToast(text,!!isErr);
    }catch(e){}
  }
  function footFetchIgoKey(){
    return fetch(footResolveHref("/igokey?format=plain"),{cache:"no-store",credentials:"same-origin"})
      .then(function(r){ if(!r.ok) throw new Error("igokey"); return r.text(); })
      .then(function(t){
        t=String(t||"").trim();
        if(!footLooksLikeIgoKey(t)) throw new Error("igokey shape");
        return t;
      });
  }
  function footCacheIgoKey(key){
    if(!footLooksLikeIgoKey(key)) return;
    igoKeyCache=key;
    igoKeyCacheAt=Date.now();
  }
  function footIgoCopyOk(){
    footHaptic();
    footIgoCopyToast("验证密钥已复制到剪贴板",false);
  }
  function footIgoCopyFail(){
    footIgoCopyToast("无法自动复制，请打开 /igokey 手动复制",true);
  }
  function footIgoCopy(){
    if(igoFetchPending) return;
    var cached=igoKeyCache;
    if(cached&&footLooksLikeIgoKey(cached)&&(Date.now()-igoKeyCacheAt)<IGO_KEY_CACHE_MS){
      footCopyPlain(cached).then(footIgoCopyOk).catch(footIgoCopyFail);
      return;
    }
    igoFetchPending=true;
    footFetchIgoKey()
      .then(function(key){
        footCacheIgoKey(key);
        if(!key) throw new Error("empty key");
        return footCopyPlain(key);
      })
      .then(footIgoCopyOk)
      .catch(footIgoCopyFail)
      .finally(function(){ igoFetchPending=false; });
  }
  footBindDoubleHit(igoHit, footIgoCopy);
  footPrefetchIgoKey=function(){
    return footFetchIgoKey().then(function(key){ footCacheIgoKey(key); return key; }).catch(function(){ return ""; });
  };
  footPrefetchIgoKey();
};
